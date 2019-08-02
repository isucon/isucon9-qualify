package main

import (
	"log"

	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/k0kubun/pp"
)

func main() {
	run()
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

	surl, err := s1.Ship(targetItemID)
	if err != nil {
		log.Fatal(err)
	}
	pp.Println(surl)

	return 0
}
