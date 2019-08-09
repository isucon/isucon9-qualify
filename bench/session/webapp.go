package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/tuotoo/qrcode"
)

type resSetting struct {
	CSRFToken string `json:"csrf_token"`
}

type resSell struct {
	ID int64 `json:"id"`
}

type reqLogin struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
}

type reqBuy struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	Token     string `json:"token"`
}

type reqSell struct {
	CSRFToken   string `json:"csrf_token"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	CategoryID  int    `json:"category_id"`
}

type reqShip struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

type resShip struct {
	Path string `json:"path"`
}

type reqBump struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

func (s *Session) Login(accountName, password string) (*asset.AppUser, error) {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, fails.NewError(err, "POST /login: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: "+msg)
	}

	u := &asset.AppUser{}
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: JSONデコードに失敗しました")
	}

	return u, nil
}

func (s *Session) SetSettings() error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/settings")
	if err != nil {
		return fails.NewError(err, "GET /settings: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "GET /settings: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(err, "GET /settings: "+msg)
	}

	rs := &resSetting{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return fails.NewError(err, "GET /settings: JSONデコードに失敗しました")
	}

	if rs.CSRFToken == "" {
		return fails.NewError(fmt.Errorf("csrf token is empty"), "GET /settings: csrf tokenが空でした")
	}

	s.csrfToken = rs.CSRFToken
	return nil
}

func (s *Session) Sell(name string, price int, description string, categoryID int) (int64, error) {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   s.csrfToken,
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, fails.NewError(err, "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, fails.NewError(err, "POST /sell: "+msg)
	}

	rs := &resSell{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return 0, fails.NewError(err, "POST /sell: JSONデコードに失敗しました")
	}

	return rs.ID, nil
}

func (s *Session) Buy(itemID int64, token string) error {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /buy: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /buy: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(err, "POST /buy: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(err, "POST /buy: bodyの読み込みに失敗しました")
	}

	return nil
}

func (s *Session) Ship(itemID int64) (apath string, err error) {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", fails.NewError(err, "POST /ship: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return "", fails.NewError(err, "POST /ship: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", fails.NewError(err, "POST /ship: "+msg)
	}

	rs := &resShip{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return "", fails.NewError(err, "POST /ship: JSONデコードに失敗しました")
	}

	if len(rs.Path) == 0 {
		return "", fails.NewError(nil, "POST /ship: Pathが空です")
	}

	return rs.Path, nil
}

func (s *Session) ShipDone(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /ship_done: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /ship_done: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(err, "POST /ship_done: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(err, "POST /ship_done: bodyの読み込みに失敗しました")
	}

	return nil
}

func (s *Session) Complete(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/complete", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /complete: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /complete: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(err, "POST /complete: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(err, "POST /complete: bodyの読み込みに失敗しました")
	}

	return nil
}

func (s *Session) DecodeQRURL(apath string) (*url.URL, error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return nil, fails.NewError(err, "リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, fails.NewError(err, "リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, fails.NewError(err, "QRコード "+msg)
	}

	qrmatrix, err := qrcode.Decode(res.Body)
	if err != nil {
		return nil, fails.NewError(err, "QRコードがデコードできませんでした")
	}

	surl := qrmatrix.Content

	if len(surl) == 0 {
		return nil, fails.NewError(nil, "QRコードの中身が空です")
	}

	sparsedURL, err := url.ParseRequestURI(surl)
	if err != nil {
		return nil, fails.NewError(err, "QRコードの中身がURLではありません")
	}

	if sparsedURL.Host != ShareTargetURLs.ShipmentURL.Host {
		return nil, fails.NewError(nil, "shipment serviceのドメイン以外のURLがQRコードに表示されています")
	}

	return sparsedURL, nil
}

func (s *Session) Bump(itemID int64) error {
	b, _ := json.Marshal(reqBump{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/bump", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(err, "POST /bump: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(err, "POST /bump: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(err, "POST /buy: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(err, "POST /bump: bodyの読み込みに失敗しました")
	}

	return nil
}
