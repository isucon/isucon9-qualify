package main

import (
	"log"
	"net/http"

	"github.com/isucon/isucon9-qualify/external/payment"
)

func main() {
	http.HandleFunc("/card", payment.CardHandler)
	http.HandleFunc("/token", payment.TokenHandler)

	log.Fatal(http.ListenAndServe(":5555", nil))
}
