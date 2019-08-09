package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/isucon/isucon9-qualify/bench/fails"
	"golang.org/x/xerrors"
)

const (
	ItemMinPrice    = 100
	ItemMaxPrice    = 1000000
	ItemPriceErrMsg = "商品価格は100円以上、1,000,000円以下にしてください"
)

type resErr struct {
	Error string `json:"error"`
}

func (s *Session) LoginWithWrongPassword(accountName, password string) error {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /login: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /login: リクエストに失敗しました")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fails.NewError(xerrors.Errorf("failed to read res.Body and the status code of the response from api was not %d: %w", http.StatusUnauthorized, err), "POST /login: bodyの読み込みに失敗しました")
		}
		return fails.NewError(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), "POST /login: 間違えたパスワードでログインした時の挙動が誤っています")
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(err, "POST /login: JSONデコードに失敗しました")
	}

	return nil
}

func (s *Session) SellWithWrongCSRFToken(name string, price int, description string, categoryID int) error {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   s.csrfToken,
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnprocessableEntity {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fails.NewError(xerrors.Errorf("failed to read res.Body and the status code of the response from api was not %d: %w", http.StatusUnprocessableEntity, err), "POST /sell: bodyの読み込みに失敗しました")
		}
		return fails.NewError(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), "POST /sell: CSRFトークンの確認が正しく動いていません")
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(err, "POST /sell: JSONデコードに失敗しました")
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
		return fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fails.NewError(xerrors.Errorf("failed to read res.Body and the status code of the response from api was not %d: %w", http.StatusBadRequest, err), "POST /sell: bodyの読み込みに失敗しました")
		}
		return fails.NewError(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), "POST /sell: 商品価格は100円以上、1,000,000円以下しか出品できません")
	}

	re := resErr{}
	err = json.NewDecoder(res.Body).Decode(&re)
	if err != nil {
		return fails.NewError(err, "POST /sell: JSONデコードに失敗しました")
	}

	if re.Error != ItemPriceErrMsg {
		return fails.NewError(err, "POST /sell: 商品価格は100円以上、1,000,000円以下しか出品できません")
	}

	return nil
}
