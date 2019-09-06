package session

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/morikuni/failure"
)

const (
	ItemMinPrice    = 100
	ItemMaxPrice    = 1000000
	ItemPriceErrMsg = "商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下にしてください"
)

type resErr struct {
	Error string `json:"error"`
}

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}

func (s *Session) LoginWithWrongPassword(ctx context.Context, accountName, password string) error {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusUnauthorized)
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /login: JSONデコードに失敗しました"))
	}

	return nil
}

func (s *Session) SellWithWrongCSRFToken(ctx context.Context, fileName, name string, price int, description string, categoryID int) error {
	file, err := os.Open(fileName)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: 画像のOpenに失敗しました"))
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "upload.jpg")
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	writer.WriteField("csrf_token", secureRandomStr(20))
	writer.WriteField("name", name)
	writer.WriteField("description", description)
	writer.WriteField("price", strconv.Itoa(price))
	writer.WriteField("category_id", strconv.Itoa(categoryID))

	contentType := writer.FormDataContentType()

	err = writer.Close()
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", contentType, body)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusUnprocessableEntity)
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: JSONデコードに失敗しました"))
	}

	return nil
}

func (s *Session) SellWithWrongPrice(ctx context.Context, fileName, name string, price int, description string, categoryID int) error {
	file, err := os.Open(fileName)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: 画像のOpenに失敗しました"))
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "upload.jpg")
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	writer.WriteField("csrf_token", s.csrfToken)
	writer.WriteField("name", name)
	writer.WriteField("description", description)
	writer.WriteField("price", strconv.Itoa(price))
	writer.WriteField("category_id", strconv.Itoa(categoryID))

	contentType := writer.FormDataContentType()

	err = writer.Close()
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", contentType, body)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusBadRequest)
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Message("POST /sell: JSONデコードに失敗しました"))
	}

	if re.Error != ItemPriceErrMsg {
		return failure.Wrap(err, failure.Message("POST /sell: 商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下しか出品できません"))
	}

	return nil
}

func (s *Session) BuyWithWrongCSRFToken(ctx context.Context, itemID int64, token string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusUnprocessableEntity, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return nil
}

func (s *Session) BuyWithFailed(ctx context.Context, itemID int64, token string, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, expectedStatus, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	if re.Error != expectedMsg {
		return failure.Wrap(err, failure.Messagef("POST /buy: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
	}

	return nil
}

func (s *Session) BuyWithFailedOnCampaign(ctx context.Context, itemID int64, token string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	// 以下の2つのどちらかのエラーになる
	if res.StatusCode == http.StatusForbidden {
		re := resErr{}
		err = json.NewDecoder(res.Body).Decode(&re)
		if err != nil {
			return failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
		}

		expectedMsg := "item is not for sale"

		if re.Error != expectedMsg {
			return failure.Wrap(err, failure.Messagef("POST /buy: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
		}

		return nil
	}

	if res.StatusCode == http.StatusBadRequest {
		re := resErr{}
		err = json.NewDecoder(res.Body).Decode(&re)
		if err != nil {
			return failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
		}

		expectedMsg := "カードの残高が足りません"

		if re.Error != expectedMsg {
			return failure.Wrap(err, failure.Messagef("POST /buy: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
		}

		return nil
	}

	return nil
}

func (s *Session) ShipWithWrongCSRFToken(ctx context.Context, itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusUnprocessableEntity, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return nil
}

func (s *Session) ShipWithFailed(ctx context.Context, itemID int64, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, expectedStatus, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	if re.Error != expectedMsg {
		return failure.Wrap(err, failure.Messagef("POST /ship: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
	}

	return nil
}

func (s *Session) DecodeQRURLWithFailed(ctx context.Context, apath string, expectedStatus int) error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, expectedStatus)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) ShipDoneWithWrongCSRFToken(ctx context.Context, itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusUnprocessableEntity, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return nil
}

func (s *Session) ShipDoneWithFailed(ctx context.Context, itemID int64, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, expectedStatus, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	if re.Error != expectedMsg {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
	}

	return nil
}

func (s *Session) ItemEditWithNotOnSale(ctx context.Context, itemID int64, price int) error {
	b, _ := json.Marshal(reqItemEdit{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		ItemPrice: price,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/items/edit", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /items/edit: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /items/edit: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusForbidden, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /items/edit: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	expectedMsg := "販売中の商品以外編集できません"

	if re.Error != expectedMsg {
		return failure.Wrap(err, failure.Messagef("POST /items/edit: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
	}

	return nil
}
