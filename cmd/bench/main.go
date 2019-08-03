package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/isucon/isucon9-qualify/external/payment"
	"github.com/isucon/isucon9-qualify/external/shipment"
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

	muxPayment := http.NewServeMux()
	muxPayment.HandleFunc("/card", payment.CardHandler)
	muxPayment.HandleFunc("/token", payment.TokenHandler)

	muxShipment := http.NewServeMux()
	muxShipment.HandleFunc("/create", shipment.CreateHandler)
	muxShipment.HandleFunc("/request", shipment.RequestHandler)
	muxShipment.HandleFunc("/accept", shipment.AcceptHandler)
	muxShipment.HandleFunc("/status", shipment.StatusHandler)

	serverPayment := &http.Server{
		Handler: &Server{
			mux: muxPayment,
		},
	}

	serverShipment := &http.Server{
		Handler: &Server{
			mux: muxShipment,
		},
	}

	go func() {
		log.Println(serverPayment.Serve(liPayment))
	}()

	go func() {
		log.Println(serverShipment.Serve(liShipment))
	}()

	run()
}

type Server struct {
	mux *http.ServeMux
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func run() int {
	err := session.SetShareTargetURLs(
		"http://localhost:8000",
		"http://localhost:5555",
		"http://localhost:7000",
	)
	if err != nil {
		log.Fatal(err)
	}

	s1, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	s2, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	seller, err := s1.Login("aaa", "aaa")
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(seller)
	err = s1.SetSettings()
	if err != nil {
		log.Fatal(err)
	}

	buyer, err := s2.Login("bbb", "bbb")
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(buyer)
	err = s2.SetSettings()
	if err != nil {
		log.Fatal(err)
	}

	targetItemID, err := s1.Sell("abcd", 100, "description description")
	if err != nil {
		log.Fatal(err)
	}
	token, err := s2.PaymentCard("AAAAAAAA", "11")
	if err != nil {
		log.Fatal(err)
	}
	err = s2.Buy(targetItemID, token)
	if err != nil {
		log.Fatal(err)
	}

	aurl, err := s1.Ship(targetItemID)
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(aurl)

	s3, err := session.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	surl, err := s3.DecodeQRURL(aurl)
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(surl.String())

	err = s3.ShipmentAccept(surl)
	if err != nil {
		log.Fatal(err)
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(6 * time.Second)

	err = s2.Complete(targetItemID)
	if err != nil {
		log.Fatal(err)
	}

	return 0
}
