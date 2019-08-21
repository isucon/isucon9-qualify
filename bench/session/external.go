package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/isucon/isucon9-qualify/bench/fails"
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

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", failure.Wrap(err, failure.Message("[payment service] /card: bodyの読み込みに失敗しました"))
		}
		return "", failure.Translate(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), fails.ErrApplication,
			failure.Messagef("[payment service] /card: got response status code %d; expected %d", res.StatusCode, http.StatusOK),
		)
	}

	rc := &resCard{}
	err = json.NewDecoder(res.Body).Decode(rc)
	if err != nil {
		return "", failure.Wrap(err, failure.Message("[payment service] /card: JSONデコードに失敗しました"))
	}

	return rc.Token, nil
}
