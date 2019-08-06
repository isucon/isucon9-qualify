package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/tuotoo/qrcode"
	"golang.org/x/xerrors"
)

type AppUser struct {
	ID          int64  `json:"id"`
	AccountName string `json:"account_name"`
	Address     string `json:"address,omitempty"`
}

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
}

type reqShip struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

type resShip struct {
	URL string `json:"url"`
}

func (s *Session) Login(accountName, password string) (*AppUser, error) {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	u := &AppUser{}
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Session) SetSettings() error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/settings")
	if err != nil {
		return err
	}

	res, err := s.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	rs := &resSetting{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return err
	}

	if rs.CSRFToken == "" {
		return fmt.Errorf("csrf token is empty")
	}

	s.csrfToken = rs.CSRFToken
	return nil
}

func (s *Session) Sell(name string, price int, description string) (int64, error) {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   s.csrfToken,
		Name:        name,
		Price:       price,
		Description: description,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return 0, xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return 0, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	rs := &resSell{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return 0, err
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
		return err
	}

	res, err := s.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) Ship(itemID int64) (aurl string, err error) {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	res, err := s.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return "", fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	rs := &resShip{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return "", err
	}

	return rs.URL, nil
}

func (s *Session) ShipDone(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	res, err := s.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
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
		return err
	}

	res, err := s.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) DecodeQRURL(aurl string) (*url.URL, error) {
	if len(aurl) == 0 {
		return nil, fmt.Errorf("urlが空です")
	}

	parsedURL, err := url.ParseRequestURI(aurl)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse url: %s: %w", aurl, err)
	}

	if parsedURL.Host != ShareTargetURLs.AppURL.Host {
		return nil, fmt.Errorf("画像はアプリケーションのドメインで配信する必要があります")
	}

	req, err := s.newGetRequest(parsedURL, "")
	if err != nil {
		return nil, err
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	qrmatrix, err := qrcode.Decode(res.Body)
	if err != nil {
		return nil, err
	}

	surl := qrmatrix.Content

	if len(surl) == 0 {
		return nil, fmt.Errorf("QRコードの中身が空です")
	}

	sparsedURL, err := url.ParseRequestURI(surl)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse url: %s: %w", surl, err)
	}

	if sparsedURL.Host != ShareTargetURLs.ShipmentURL.Host {
		return nil, fmt.Errorf("shipment serviceのドメイン以外のURLがQRコードに表示されています")
	}

	return sparsedURL, nil
}
