package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/isucon/isucon9-qualify/bench/server"
)

func main() {
	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 7000})
	if err != nil {
		log.Fatal(err)
	}

	ship := server.NewShipment()
	serverShipment := &http.Server{
		Handler: ship,
	}

	ship.SetDelay(200 * time.Millisecond)

	log.Print(serverShipment.Serve(liShipment))
}
