package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0"

	userAgent = "isucon9-qualify-webapp"
)

type APIPaymentServiceTokenReq struct {
	ShopID string `json:"shop_id"`
	Token  string `json:"token"`
	APIKey string `json:"api_key"`
	Price  int    `json:"price"`
}

type APIPaymentServiceTokenRes struct {
	Status string `json:"status"`
}

type APIShipmentCreateReq struct {
	ToAddress   string `json:"to_address"`
	ToName      string `json:"to_name"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

type APIShipmentCreateRes struct {
	ReserveID   string `json:"reserve_id"`
	ReserveTime int64  `json:"reserve_time"`
}

type APIShipmentRequestReq struct {
	ReserveID string `json:"reserve_id"`
}

type APIShipmentStatusRes struct {
	Status      string `json:"status"`
	ReserveTime int64  `json:"reserve_time"`
}

type APIShipmentStatusReq struct {
	ReserveID string `json:"reserve_id"`
}

func APIPaymentToken(paymentURL string, param *APIPaymentServiceTokenReq) (*APIPaymentServiceTokenRes, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodPost, paymentURL+"/token", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
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

	pstr := &APIPaymentServiceTokenRes{}
	err = json.NewDecoder(res.Body).Decode(pstr)
	if err != nil {
		return nil, err
	}

	return pstr, nil
}

func APIShipmentCreate(shipmentURL string, param *APIShipmentCreateReq) (*APIShipmentCreateRes, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodPost, shipmentURL+"/create", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
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

	scr := &APIShipmentCreateRes{}
	err = json.NewDecoder(res.Body).Decode(&scr)
	if err != nil {
		return nil, err
	}

	return scr, nil
}

func APIShipmentRequest(shipmentURL string, param *APIShipmentRequestReq) ([]byte, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodPost, shipmentURL+"/request", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
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

	return ioutil.ReadAll(res.Body)
}

func APIShipmentStatus(shipmentURL string, param *APIShipmentStatusReq) (*APIShipmentStatusRes, error) {
	b, _ := json.Marshal(param)

	req, err := http.NewRequest(http.MethodGet, shipmentURL+"/status", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
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

	ssr := &APIShipmentStatusRes{}
	err = json.NewDecoder(res.Body).Decode(&ssr)
	if err != nil {
		return nil, err
	}

	return ssr, nil
}
