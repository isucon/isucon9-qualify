package main

import (
	"log"
	"net/http"

	"github.com/isucon/isucon9-qualify/external/shipment"
)

func main() {
	http.HandleFunc("/create", shipment.CreateHandler)
	http.HandleFunc("/request", shipment.RequestHandler)
	http.HandleFunc("/accept", shipment.AcceptHandler)
	http.HandleFunc("/status", shipment.StatusHandler)

	log.Fatal(http.ListenAndServe(":7000", nil))
}
