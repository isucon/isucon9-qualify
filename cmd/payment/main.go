package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/isucon/isucon9-qualify/bench/server"
)

func main() {
	liPayment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 5555})
	if err != nil {
		log.Fatal(err)
	}

	pay := server.NewPayment(nil)

	serverPayment := &http.Server{
		Handler: pay,
	}

	pay.SetDelay(200 * time.Millisecond)

	log.Print(serverPayment.Serve(liPayment))
}
