package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/isucon/isucon9-qualify/bench/server"
)

func main() {
	flags := flag.NewFlagSet("shipment", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	dataDir := ""
	port := 0

	flags.StringVar(&dataDir, "data-dir", "initial-data", "data directory")
	flags.IntVar(&port, "port", 7001, "shipment service port")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
	if err != nil {
		log.Fatal(err)
	}

	ship := server.NewShipment(true, dataDir, nil)
	serverShipment := &http.Server{
		Handler: ship,
	}

	ship.SetDelay(200 * time.Millisecond)

	log.Print(serverShipment.Serve(liShipment))
}
