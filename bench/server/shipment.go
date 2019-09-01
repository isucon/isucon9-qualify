package server

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
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

type AppShipping struct {
	TransactionEvidenceID int64  `json:"transaction_evidence_id" db:"transaction_evidence_id"`
	Status                string `json:"status" db:"status"`
	ReserveID             string `json:"reserve_id" db:"reserve_id"`
	ReserveTime           int64  `json:"reserve_time" db:"reserve_time"`
	ToAddress             string `json:"to_address" db:"to_address"`
	ToName                string `json:"to_name" db:"to_name"`
	FromAddress           string `json:"from_address" db:"from_address"`
	FromName              string `json:"from_name" db:"from_name"`
}

type shipment struct {
	ToAddress   string `json:"to_address"`
	ToName      string `json:"to_name"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`

	Status          string    `json:"-"`
	QRMD5           string    `json:"-"`
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

	c.items[key] = value

	return value, true
}

func (c *shipmentStore) SetQRMD5(key string, str string) (shipment, bool) {
	c.Lock()
	defer c.Unlock()

	value, ok := c.items[key]
	if !ok {
		return shipment{}, false
	}
	value.QRMD5 = str

	c.items[key] = value

	return value, true
}

func (c *shipmentStore) SetStatusWithDone(key string, doneDatetime time.Time) (shipment, bool) {
	c.Lock()
	defer c.Unlock()

	value, ok := c.items[key]
	if !ok {
		return shipment{}, false
	}
	value.Status = StatusShipping
	value.DoneDatetime = doneDatetime

	c.items[key] = value

	return value, true
}

func (c *shipmentStore) ForceSet(key string, value shipment) {
	c.Lock()
	c.items[key] = value
	c.Unlock()
}

func (c *shipmentStore) Get(key string) (shipment, bool) {
	c.Lock()
	defer c.Unlock()

	v, found := c.items[key]
	if v.Status == StatusShipping && !v.DoneDatetime.IsZero() && time.Now().After(v.DoneDatetime) {
		v.Status = StatusDone
	}

	return v, found
}

func init() {
	rand.Seed(time.Now().UnixNano())

	shipmentHash = sha1.New()
	shipmentHash.Write(SecretSeed)
}

type createRes struct {
	ReserveID   string `json:"reserve_id"`
	ReserveTime int64  `json:"reserve_time"`
}

type ServerShipment struct {
	debug         bool
	shipmentCache *shipmentStore

	Server
}

func NewShipment(debug bool, dataDir string, allowedIPs []net.IP) *ServerShipment {
	s := &ServerShipment{
		debug: debug,
	}

	s.shipmentCache = NewShipmentStore()

	f, err := os.Open(filepath.Join(dataDir, "result/shippings_json.txt"))
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	ship := AppShipping{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), &ship)
		if err != nil {
			log.Fatal(err)
		}
		s.shipmentCache.ForceSet(ship.ReserveID, shipment{
			ToAddress:       ship.ToAddress,
			ToName:          ship.ToName,
			FromAddress:     ship.FromAddress,
			FromName:        ship.FromName,
			Status:          ship.Status,
			ReserveDatetime: time.Unix(ship.ReserveTime, 0),
		})
	}
	f.Close()

	s.mux = http.NewServeMux()
	s.allowedIPs = allowedIPs

	s.mux.Handle("/create", apply(http.HandlerFunc(s.createHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/request", apply(http.HandlerFunc(s.requestHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/accept", apply(http.HandlerFunc(s.acceptHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/status", apply(http.HandlerFunc(s.statusHandler), s.withDelay(), s.withIPRestriction()))

	return s
}

func (s *ServerShipment) createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

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
	res.ReserveID = s.shipmentCache.Set(ship)
	res.ReserveTime = ship.ReserveDatetime.Unix()

	json.NewEncoder(w).Encode(res)
}

type requestReq struct {
	ReserveID string `json:"reserve_id"`
}

func (s *ServerShipment) requestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

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

	_, ok := s.shipmentCache.SetStatus(req.ReserveID, StatusWaitPickup)
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

	if s.debug {
		log.Print(msg)
	}

	png, err := qrcode.Encode(msg, qrcode.Low, 256)
	if err != nil {
		log.Print(err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h := md5.New()
	h.Write(png)

	s.shipmentCache.SetQRMD5(req.ReserveID, fmt.Sprintf("%x", h.Sum(nil)))

	w.Header().Set("Content-Type", "image/png")

	w.Write(png)
}

func (s *ServerShipment) acceptHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	token := query.Get("token")

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	if token != fmt.Sprintf("%x", shipmentHash.Sum([]byte(id))) {
		b, _ := json.Marshal(errorRes{Error: "wrong parameters"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}

	_, ok := s.shipmentCache.SetStatusWithDone(id, time.Now().Add(5*time.Second))
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

func (s *ServerShipment) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != IsucariAPIToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

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

	ship, ok := s.shipmentCache.Get(req.ReserveID)
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

func (s *ServerShipment) ForceSetStatus(key string, status string) bool {
	_, ok := s.shipmentCache.SetStatus(key, status)

	return ok
}

func (s *ServerShipment) CheckQRMD5(key string, md5Str string) bool {
	val, ok := s.shipmentCache.Get(key)
	if !ok {
		return false
	}

	return val.QRMD5 == md5Str
}
