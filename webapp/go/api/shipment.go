package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0"
)

type ShipmentCreateReq struct {
	ToAddress   string `json:"to_address"`
	ToName      string `json:"to_name"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

type ShipmentCreateRes struct {
	ReserveID   string `json:"reserve_id"`
	ReserveTime int64  `json:"reserve_time"`
}

func ShipmentCreate(shipmentURL string, param *ShipmentCreateReq) (*ShipmentCreateRes, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodPost, shipmentURL+"/create", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", IsucariAPIToken)

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

	scr := &ShipmentCreateRes{}
	err = json.NewDecoder(res.Body).Decode(&scr)
	if err != nil {
		return nil, err
	}

	return scr, nil
}
