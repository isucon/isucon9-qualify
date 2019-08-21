package session

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/morikuni/failure"
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
		return "", failure.Wrap(err, failure.Message("[payment service] /card: リクエストに失敗しました"))
	}

	req.Header.Add("Origin", "http://localhost:8000")

	res, err := s.Do(req)
	if err != nil {
		return "", failure.Wrap(err, failure.Message("[payment service] /card: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", failure.Wrap(err, failure.Message("[payment service] /card: "+msg))
	}

	rc := &resCard{}
	err = json.NewDecoder(res.Body).Decode(rc)
	if err != nil {
		return "", failure.Wrap(err, failure.Message("[payment service] /card: JSONデコードに失敗しました"))
	}

	return rc.Token, nil
}
