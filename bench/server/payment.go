package server

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/morikuni/failure"
)

const (
	IsucariAPIKey = "a15400e46c83635eb181-946abb51ff26a868317c"
	IsucariShopID = "11"
)

var (
	regex = regexp.MustCompile("^[0-9A-F]{8}$")
)

type cardReq struct {
	CardNumber string `json:"card_number"`
	ShopID     string `json:"shop_id"`
}

type cardRes struct {
	Token string `json:"token"`
}

type tokenReq struct {
	ShopID string `json:"shop_id"`
	Token  string `json:"token"`
	APIKey string `json:"api_key"`
	Price  int    `json:"price"`
}

type tokenRes struct {
	Status string `json:"status"`
}

type cardTokenStore struct {
	sync.Mutex
	items map[string]cardToken
}

type cardToken struct {
	number string
	expire time.Time

	// for benchmarker
	itemID int64
	price  int
}

func newCardToken() *cardTokenStore {
	m := make(map[string]cardToken)
	c := &cardTokenStore{
		items: m,
	}
	return c
}

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}

func (c *cardTokenStore) Set(card string) string {
	token := secureRandomStr(20)
	expire := time.Now().Add(5 * time.Minute)
	c.Lock()
	c.items[token] = cardToken{
		number: card,
		expire: expire,
	}
	c.Unlock()

	return token
}

func (c *cardTokenStore) Get(token string) (cardToken, bool) {
	c.Lock()
	v, found := c.items[token]
	delete(c.items, token)
	c.Unlock()

	if time.Now().After(v.expire) {
		return cardToken{}, false
	}

	return v, found
}

func newReports() *reportStore {
	m := make(map[int64]report)
	c := &reportStore{
		items: m,
	}
	return c
}

type reportStore struct {
	sync.Mutex
	items map[int64]report
}

type report struct {
	Price  int
	Status string
}

func (c *reportStore) Set(itemID int64, price int) {
	c.Lock()
	defer c.Unlock()

	_, ok := c.items[itemID]
	if ok {
		fails.ErrorsForCheck.Add(failure.New(fails.ErrCritical, failure.Messagef("多重決済を検知しました (item_id: %d)", itemID)))
		return
	}

	c.items[itemID] = report{
		Price: price,
		// statusがdoneになったかどうかだけを確認しているので、初期化時は特に必要ない
		// Status: asset.TransactionEvidenceStatusWaitShipping,
	}
}

func (c *reportStore) SetStatus(itemID int64, status string) {
	c.Lock()
	item := c.items[itemID]
	item.Status = status
	c.items[itemID] = item
	c.Unlock()
}

type ServerPayment struct {
	cardTokens *cardTokenStore
	reports    *reportStore

	Server
}

func NewPayment(allowedIPs []net.IP) *ServerPayment {
	s := &ServerPayment{}

	s.cardTokens = newCardToken()
	s.reports = newReports()
	s.mux = http.NewServeMux()
	s.allowedIPs = allowedIPs

	s.mux.Handle("/card", apply(http.HandlerFunc(s.cardHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/token", apply(http.HandlerFunc(s.tokenHandler), s.withDelay(), s.withIPRestriction()))

	return s
}

func (s *ServerPayment) tokenHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	tr := tokenReq{}
	err := json.NewDecoder(req.Body).Decode(&tr)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if tr.ShopID != IsucariShopID {
		b, _ := json.Marshal(errorRes{Error: "wrong shop id"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if tr.APIKey != IsucariAPIKey {
		b, _ := json.Marshal(errorRes{Error: "wrong api key"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	ct, ok := s.cardTokens.Get(tr.Token)
	if !ok {
		result := tokenRes{
			Status: "invalid",
		}

		b, _ := json.Marshal(result)

		w.Write(b)
		return
	}

	if strings.Contains(ct.number, "FA10") {
		result := tokenRes{
			Status: "fail",
		}

		b, _ := json.Marshal(result)

		w.Write(b)
		return
	}

	result := tokenRes{
		Status: "ok",
	}

	if ct.price != 0 {
		if ct.price != tr.Price {
			// エラーにはするが処理を継続する
			fails.ErrorsForCheck.Add(failure.New(fails.ErrCritical, failure.Messagef("決済額に誤りがあります expected: %d; actual: %d", ct.price, tr.Price)))

			b, _ := json.Marshal(result)

			w.WriteHeader(http.StatusForbidden)
			w.Write(b)
			return
		}

		s.reports.Set(ct.itemID, ct.price)
	}

	json.NewEncoder(w).Encode(result)
}

func isValidOrigin(origin string) bool {
	return true
}

func (s *ServerPayment) cardHandler(w http.ResponseWriter, req *http.Request) {
	if !isValidOrigin(req.Header.Get("Origin")) {
		return
	}

	// Originはちゃんとチェックしている前提のコード。コピペしないこと。
	w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	if req.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	cr := cardReq{}
	err := json.NewDecoder(req.Body).Decode(&cr)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if cr.ShopID != IsucariShopID {
		b, _ := json.Marshal(errorRes{Error: "wrong shop id"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	if !regex.MatchString(cr.CardNumber) {
		b, _ := json.Marshal(errorRes{Error: "card number is wrong"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	token := s.cardTokens.Set(cr.CardNumber)

	res := cardRes{
		Token: token,
	}

	json.NewEncoder(w).Encode(res)
}

// ForceSet is the function for benchmarker
func (s *ServerPayment) ForceSet(card string, itemID int64, price int) string {
	token := secureRandomStr(20)
	expire := time.Now().Add(5 * time.Minute)
	s.cardTokens.Lock()
	s.cardTokens.items[token] = cardToken{
		number: card,
		expire: expire,
		itemID: itemID,
		price:  price,
	}
	s.cardTokens.Unlock()

	return token
}

// ForceReportsSetStatus is the function for benchmarker
func (s *ServerPayment) ForceReportsSetStatus(itemID int64, status string) {
	s.reports.SetStatus(itemID, status)
}

// GetReports is the function for benchmarker
// コピーはしていないので注意
func (s *ServerPayment) GetReports() map[int64]report {
	s.reports.Lock()
	defer s.reports.Unlock()

	return s.reports.items
}
