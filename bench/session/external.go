package session

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/isucon/isucon9-qualify/bench/fails"
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
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "[payment service] /card: リクエストに失敗しました")
	}

	req.Header.Add("Origin", "http://localhost:8000")

	res, err := s.Do(req)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "[payment service] /card: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "[payment service] /card: "+msg)
	}

	rc := &resCard{}
	err = json.NewDecoder(res.Body).Decode(rc)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "[payment service] /card: JSONデコードに失敗しました")
	}

	return rc.Token, nil
}

func (s *Session) ShipmentAccept(surl *url.URL) error {
	req, err := s.newGetRequest(*surl, "")
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "[shipment service] /accept: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "[shipment service] /accept: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "[shipment service] /accept: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "[shipment service] /accept: bodyの読み込みに失敗しました")
	}

	return nil
}
