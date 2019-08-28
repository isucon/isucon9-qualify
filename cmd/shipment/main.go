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

	flags.StringVar(&dataDir, "data-dir", "initial-data", "data directory")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 7000})
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
