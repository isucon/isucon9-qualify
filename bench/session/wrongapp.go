package session

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/isucon/isucon9-qualify/bench/fails"
	"golang.org/x/xerrors"
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

func (s *Session) LoginWithWrongPassword(accountName, password string) error {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /login: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /login: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusUnauthorized)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /login: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /login: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) SellWithWrongCSRFToken(name string, price int, description string, categoryID int) error {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   secureRandomStr(20),
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusUnprocessableEntity)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) SellWithWrongPrice(name string, price int, description string, categoryID int) error {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   s.csrfToken,
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusBadRequest)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: JSONデコードに失敗しました")
	}

	if re.Error != ItemPriceErrMsg {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: 商品価格は100ｲｽｺｲﾝ以上、1,000,000ｲｽｺｲﾝ以下しか出品できません")
	}

	return nil
}

func (s *Session) BuyWithWrongCSRFToken(itemID int64, token string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusUnprocessableEntity)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) BuyWithFailed(itemID int64, token string, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, expectedStatus)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: JSONデコードに失敗しました")
	}

	if re.Error != expectedMsg {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("POST /buy: exected error message: %s; actual: %s", expectedMsg, re.Error))
	}

	return nil
}

func (s *Session) ShipWithWrongCSRFToken(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusUnprocessableEntity)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) ShipWithFailed(itemID int64, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, expectedStatus)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: JSONデコードに失敗しました")
	}

	if re.Error != expectedMsg {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("POST /ship: exected error message: %s; actual: %s", expectedMsg, re.Error))
	}

	return nil
}

func (s *Session) DecodeQRURLWithFailed(apath string, expectedStatus int) error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: リクエストに失敗しました", apath))
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, expectedStatus)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: %s", apath, msg))
	}

	return nil
}

func (s *Session) ShipDoneWithWrongCSRFToken(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: secureRandomStr(20),
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusUnprocessableEntity)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) ShipDoneWithFailed(itemID int64, expectedStatus int, expectedMsg string) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, expectedStatus)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: "+msg)
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: JSONデコードに失敗しました")
	}

	if re.Error != expectedMsg {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("POST /ship_done: exected error message: %s; actual: %s", expectedMsg, re.Error))
	}

	return nil
}
