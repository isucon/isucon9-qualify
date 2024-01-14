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
	flags := flag.NewFlagSet("payment", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	port := 0

	flags.IntVar(&port, "port", 5555, "payment service port")
	err := flags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	liPayment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: port})
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
