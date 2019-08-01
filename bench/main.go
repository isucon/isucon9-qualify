package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tuotoo/qrcode"
)

const (
	UserAgent = "benchmarker"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type resSell struct {
	ID int64 `json:"id" db:"id"`
}

type cardReq struct {
	CardNumber string `json:"card_number"`
	ShopID     string `json:"shop_id"`
}

type cardRes struct {
	Token string `json:"token"`
}

type reqBuy struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	Token     string `json:"token"`
}

func main() {
	s1 := NewSession()
	s2 := NewSession()

	formData := url.Values{}
	formData.Set("account_name", "aaa")
	formData.Set("password", "aaa")
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8000/login", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := s1.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to login")
	}

	formData = url.Values{}
	formData.Set("account_name", "bbb")
	formData.Set("password", "bbb")
	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/login", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = s2.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to login")
	}

	req, err = http.NewRequest(http.MethodGet, "http://localhost:8000/sell", nil)
	res, err = s1.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	csrfToken1, ok := doc.Find(`input[name="csrf_token"]`).First().Attr("value")
	if !ok {
		log.Fatal("cannot get csrf token")
	}

	formData = url.Values{}
	formData.Set("csrf_token", csrfToken1)
	formData.Set("name", "test1")
	formData.Set("price", "300")
	formData.Set("description", "testtest")

	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/sell", bytes.NewBufferString(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err = s1.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to sell")
	}

	rs := resSell{}
	err = json.NewDecoder(res.Body).Decode(&rs)
	if err != nil {
		log.Fatal(err)
	}

	targetItemID := rs.ID

	req, err = http.NewRequest(http.MethodGet, "http://localhost:8000/sell", nil)
	res, err = s2.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	csrfToken2, ok := doc.Find(`input[name="csrf_token"]`).First().Attr("value")
	if !ok {
		log.Fatal("cannot get csrf token")
	}

	creq := cardReq{
		CardNumber: "AAAAAAAA",
		ShopID:     "11",
	}
	b, _ := json.Marshal(creq)
	req, err = http.NewRequest(http.MethodPost, "http://localhost:5555/card", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Origin", "http://localhost:8000")
	res, err = s2.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}

	cres := cardRes{}
	err = json.NewDecoder(res.Body).Decode(&cres)
	if err != nil {
		log.Fatal(err)
	}

	b, _ = json.Marshal(reqBuy{
		CSRFToken: csrfToken2,
		ItemID:    targetItemID,
		Token:     cres.Token,
	})

	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/buy", bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}

	res, err = s2.SendRequest(req)
	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to buy")
	}

	formData = url.Values{}
	formData.Set("csrf_token", csrfToken1)
	formData.Set("item_id", fmt.Sprintf("%d", targetItemID))
	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/ship", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = s1.SendRequest(req)
	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to ship")
	}

	shipRes := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&shipRes)
	if err != nil {
		log.Fatal(err)
	}
	if shipRes["url"] == "" {
		log.Fatal("url is empty")
	}

	res, err = http.Get(shipRes["url"])
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	qrmatrix, err := qrcode.Decode(res.Body)
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = http.Get(qrmatrix.Content)
	if err != nil {
		log.Fatal(err)
		return
	}

	formData = url.Values{}
	formData.Set("csrf_token", csrfToken1)
	formData.Set("item_id", fmt.Sprintf("%d", targetItemID))
	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/ship_done", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = s1.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatal("failed to ship_done")
	}

	time.Sleep(6 * time.Second)

	formData = url.Values{}
	formData.Set("csrf_token", csrfToken2)
	formData.Set("item_id", fmt.Sprintf("%d", targetItemID))
	req, err = http.NewRequest(http.MethodPost, "http://localhost:8000/complete", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = s2.SendRequest(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(b))
		log.Fatal("failed to complete")
	}
}

type Session struct {
	Client *http.Client
}

func NewSession() *Session {
	s := &Session{}

	jar, _ := cookiejar.New(&cookiejar.Options{})
	s.Client = &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("redirect attempted")
		},
	}

	return s
}

func (s *Session) SendRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)

	return s.Client.Do(req)
}
