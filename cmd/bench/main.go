package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/k0kubun/pp"
)

func main() {
	liPayment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 5555})
	if err != nil {
		log.Fatal(err)
	}

	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 7000})
	if err != nil {
		log.Fatal(err)
	}

	serverPayment := &http.Server{
		Handler: server.NewPayment(),
	}

	serverShipment := &http.Server{
		Handler: server.NewShipment(),
	}

	go func() {
		log.Println(serverPayment.Serve(liPayment))
	}()

	go func() {
		log.Println(serverShipment.Serve(liShipment))
	}()

	fmt.Fprintf(os.Stderr, "=== initialize ===\n")
	initialize()
	fmt.Fprintf(os.Stderr, "=== verify ===\n")

	cerr := verify()
	criticalMsgs := cerr.GetMsgs()
	if len(criticalMsgs) > 0 {
		fmt.Fprintf(os.Stderr, "cause error!\n")

		output := Output{
			Pass:     false,
			Score:    0,
			Messages: criticalMsgs,
		}
		json.NewEncoder(os.Stdout).Encode(output)

		return
	}

	fmt.Fprintf(os.Stderr, "=== validation ===\n")
}

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Messages []string `json:"messages"`
}

type Server struct {
	mux *http.ServeMux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func initialize() {
}

func scenarioSellAndBuy() error {
	err := session.SetShareTargetURLs(
		"http://localhost:8000",
		"http://localhost:5555",
		"http://localhost:7000",
	)
	if err != nil {
		return err
	}

	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	s2, err := session.NewSession()
	if err != nil {
		return err
	}

	seller, err := s1.Login("aaa", "aaa")
	if err != nil {
		return err
	}
	pp.Println(seller)
	err = s1.SetSettings()
	if err != nil {
		return err
	}

	buyer, err := s2.Login("bbb", "bbb")
	if err != nil {
		return err
	}
	pp.Println(buyer)
	err = s2.SetSettings()
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}
	token, err := s2.PaymentCard("AAAAAAAA", "11")
	if err != nil {
		return err
	}
	err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	aurl, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	s3, err := session.NewSession()
	if err != nil {
		return err
	}
	surl, err := s3.DecodeQRURL(aurl)
	if err != nil {
		return err
	}
	pp.Println(surl.String())

	err = s3.ShipmentAccept(surl)
	if err != nil {
		return err
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	time.Sleep(6 * time.Second)

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}

func verify() *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := scenarioSellAndBuy()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}
