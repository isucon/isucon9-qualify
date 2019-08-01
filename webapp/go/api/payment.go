package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PaymentServiceTokenReq struct {
	Token  string `json:"token"`
	APIKey string `json:"api_key"`
	Price  int    `json:"price"`
}

type PaymentServiceTokenRes struct {
	Status string `json:"status"`
}

func PaymentToken(paymentURL string, param *PaymentServiceTokenReq) (*PaymentServiceTokenRes, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodPost, paymentURL+"/token", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read res.Body and the status code of the response from shipment service was not 200: %v", err)
		}
		return nil, fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	pstr := &PaymentServiceTokenRes{}
	err = json.NewDecoder(res.Body).Decode(pstr)
	if err != nil {
		return nil, err
	}

	return pstr, nil
}
