package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/isucon/isucon9-qualify/bench/scenario"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
)

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Messages []string `json:"messages"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	err = session.SetShareTargetURLs(
		"http://localhost:8000",
		"http://localhost:5555",
		"http://localhost:7000",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "=== initialize ===\n")
	scenario.Initialize()
	fmt.Fprintf(os.Stderr, "=== verify ===\n")

	cerr := scenario.Verify()
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

	output := Output{
		Pass:  true,
		Score: 0,
	}
	json.NewEncoder(os.Stdout).Encode(output)
}
