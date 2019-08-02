package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/xerrors"
)

type reqCard struct {
	CardNumber string `json:"card_number"`
	ShopID     string `json:"shop_id"`
}

type resCard struct {
	Token string `json:"token"`
}

func (s *Session) PaymentCard(cardNumber, shopID string) (token string, err error) {
	b, _ := json.Marshal(reqCard{
		CardNumber: cardNumber,
		ShopID:     shopID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.PaymentURL, "/card", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	req.Header.Add("Origin", "http://localhost:8000")

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

	rc := &resCard{}
	err = json.NewDecoder(res.Body).Decode(rc)
	if err != nil {
		return "", err
	}

	return rc.Token, nil
}
