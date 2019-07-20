package main

import (
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	IsucariAPIKey = "a15400e46c83635eb181-946abb51ff26a868317c"
	IsucariShopID = "11"
)

type cardReq struct {
	CardNumber string `json:"card_number"`
	ShopID     string `json:"shop_id"`
}

type cardRes struct {
	Token string `json:"token"`
}

type tokenReq struct {
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
}

type errorRes struct {
	Error string `json:"error"`
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

var CardTokens = newCardToken()

func main() {
	http.HandleFunc("/card", cardHandler)
	http.HandleFunc("/token", tokenHandler)

	log.Fatal(http.ListenAndServe(":5555", nil))
}

func tokenHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tr := tokenReq{}
	err := json.NewDecoder(req.Body).Decode(&tr)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b)

		return
	}

	if tr.APIKey != IsucariAPIKey {
		b, _ := json.Marshal(errorRes{Error: "wrong api key"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	_, ok := CardTokens.Get(tr.Token)
	if !ok {
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

	b, _ := json.Marshal(result)

	w.Write(b)
}

func isValidOrigin(origin string) bool {
	return true
}

func cardHandler(w http.ResponseWriter, req *http.Request) {
	if !isValidOrigin(req.Header.Get("Origin")) {
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "http://localhost:8000")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	if req.Method == http.MethodOptions {
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	cr := cardReq{}
	err := json.NewDecoder(req.Body).Decode(&cr)
	if err != nil {
		b, _ := json.Marshal(errorRes{Error: "json decode error"})

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b)

		return
	}

	if cr.ShopID != IsucariShopID {
		b, _ := json.Marshal(errorRes{Error: "wrong shop id"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)

		return
	}

	token := CardTokens.Set(cr.CardNumber)

	res := cardRes{
		Token: token,
	}

	b, _ := json.Marshal(res)

	w.Write(b)
}
