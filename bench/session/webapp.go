package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/k0kubun/pp"
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
	pp.Println(s.csrfToken)
	return nil
}

func (s *Session) Sell(name string, price int, description string) (int64, error) {
	formData := url.Values{}
	formData.Set("csrf_token", s.csrfToken)
	formData.Set("name", name)
	formData.Set("price", strconv.Itoa(price))
	formData.Set("description", description)

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
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
