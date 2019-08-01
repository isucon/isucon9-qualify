package session

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/xerrors"
)

func (s *Session) Login(accountName, password string) error {
	formData := url.Values{}
	formData.Set("account_name", accountName)
	formData.Set("password", password)

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/x-www-form-urlencoded", bytes.NewBufferString(formData.Encode()))
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
			return nil, xerrors.Errorf("failed to read res.Body and the status code of the response from api was not 200: %w", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}
