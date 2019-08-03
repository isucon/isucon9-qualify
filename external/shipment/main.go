package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"image/png"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

var (
	SecretSeed   = []byte("secret-seed")
	shipmentHash hash.Hash
)

const (
	StatusInitial    = "initial"
	StatusWaitPickup = "wait_pickup"
	StatusShipping   = "shipping"
	StatusDone       = "done"

	IsucariAPIToken = "Bearer 75ugk2m37a750fwir5xr-22l6h4wmue1bwrubzwd0"
)

type shipment struct {
	ToAddress   string `json:"to_address"`
	ToName      string `json:"to_name"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`

	Status          string    `json:"-"`
	ReserveDatetime time.Time `json:"-"`
	DoneDatetime    time.Time `json:"-"`
}

type shipmentStatusRes struct {
	Status      string `json:"status"`
	ReserveTime int64  `json:"reserve_time"`
}

type shipmentStatusReq struct {
	ReserveID string `json:"reserve_id"`
}

type shipmentStore struct {
	sync.Mutex
	items map[string]shipment
}

func NewShipmentStore() *shipmentStore {
	m := make(map[string]shipment)
	c := &shipmentStore{
		items: m,
	}
	return c
}

func (c *shipmentStore) Set(value shipment) string {
	key := ""

	c.Lock()
	for ok := true; ok; {
		key = fmt.Sprintf("%010d", rand.Intn(10000000000))
		_, ok = c.items[key]
	}
	c.items[key] = value
	c.Unlock()

	return key
}

func (c *shipmentStore) SetStatus(key string, status string) (shipment, bool) {
	c.Lock()
	defer c.Unlock()

	value, ok := c.items[key]
	if !ok {
		return shipment{}, false
	}
	value.Status = status
	if status == StatusShipping {
		value.DoneDatetime = time.Now().Add(5 * time.Second)
	}

	c.items[key] = value

	return value, true
}

func (c *shipmentStore) Get(key string) (shipment, bool) {
	c.Lock()
	defer c.Unlock()

	v, found := c.items[key]
	if v.Status == StatusShipping && time.Now().After(v.DoneDatetime) {
		v.Status = StatusDone
	}

	return v, found
}

var shipmentCache = NewShipmentStore()

func init() {
	rand.Seed(time.Now().UnixNano())

	shipmentHash = sha1.New()
	shipmentHash.Write(SecretSeed)
}

func main() {
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/request", requestHandler)
	http.HandleFunc("/accept", acceptHandler)
	http.HandleFunc("/status", statusHandler)

	http.ListenAndServe(":7000", nil)
}

type errorRes struct {
	Error string `json:"error"`
}

type createRes struct {
	ReserveID   string `json:"reserve_id"`
	ReserveTime int64  `json:"reserve_time"`
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ship := shipment{}
	err := json.NewDecoder(r.Body).Decode(&ship)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if ship.ToAddress == "" || ship.ToName == "" || ship.FromAddress == "" || ship.FromName == "" {
		b, _ := json.Marshal(errorRes{Error: "required parameter was not passed"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	now := time.Now()
	ship.ReserveDatetime = now
	ship.Status = StatusInitial

	res := createRes{}
	res.ReserveID = shipmentCache.Set(ship)
	res.ReserveTime = ship.ReserveDatetime.Unix()

	json.NewEncoder(w).Encode(res)
}

type requestReq struct {
	ReserveID string `json:"reserve_id"`
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := requestReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if req.ReserveID == "" {
		b, _ := json.Marshal(errorRes{Error: "required parameter was not passed"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	_, ok := shipmentCache.SetStatus(req.ReserveID, StatusWaitPickup)
	if !ok {
		b, _ := json.Marshal(errorRes{Error: "empty"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   "/accept",
	}
	q := u.Query()
	q.Set("id", req.ReserveID)
	q.Set("token", fmt.Sprintf("%x", shipmentHash.Sum([]byte(req.ReserveID))))

	u.RawQuery = q.Encode()

	msg := u.String()
	fmt.Println(msg)

	qrCode, _ := qr.Encode(msg, qr.L, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, 256, 256)

	png.Encode(w, qrCode)
}

func acceptHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	token := query.Get("token")

	if token != fmt.Sprintf("%x", shipmentHash.Sum([]byte(id))) {
		b, _ := json.Marshal(errorRes{Error: "wrong parameters"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	_, ok := shipmentCache.SetStatus(id, StatusShipping)
	if !ok {
		b, _ := json.Marshal(errorRes{Error: "empty"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	b, _ := json.Marshal(struct {
		Accept string `json:"accept"`
	}{
		Accept: "ok",
	})
	w.Write(b)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	req := shipmentStatusReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if req.ReserveID == "" {
		b, _ := json.Marshal(errorRes{Error: "required parameter was not passed"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
	}

	ship, ok := shipmentCache.Get(req.ReserveID)
	if !ok {
		b, _ := json.Marshal(errorRes{Error: "empty"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	res := shipmentStatusRes{}
	res.Status = ship.Status
	res.ReserveTime = ship.ReserveDatetime.Unix()

	json.NewEncoder(w).Encode(res)
}
