package server

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/isucon/isucon9-qualify/external/payment"
	"github.com/isucon/isucon9-qualify/external/shipment"
)

type Server struct {
	mux *http.ServeMux
}

type Adapter func(http.Handler) http.Handler

func NewShipment() *Server {
	s := &Server{}

	s.mux = http.NewServeMux()

	s.mux.HandleFunc("/create", shipment.CreateHandler)
	s.mux.HandleFunc("/request", shipment.RequestHandler)
	s.mux.HandleFunc("/accept", shipment.AcceptHandler)
	s.mux.HandleFunc("/status", shipment.StatusHandler)

	return s
}

func NewPayment() *Server {
	s := &Server{}

	s.mux = http.NewServeMux()

	s.mux.Handle("/card", apply(http.HandlerFunc(payment.CardHandler), withIPRestriction()))
	s.mux.Handle("/token", apply(http.HandlerFunc(payment.TokenHandler), withIPRestriction()))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func tmp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func withIPRestriction() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

func apply(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adpt := range adapters {
		h = adpt(h)
	}
	return h
}

func RunServer(paymentPort, shipmentPort int) error {
	liPayment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: paymentPort})
	if err != nil {
		return err
	}

	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: shipmentPort})
	if err != nil {
		return err
	}

	serverPayment := &http.Server{
		Handler: NewPayment(),
	}

	serverShipment := &http.Server{
		Handler: NewShipment(),
	}

	go func() {
		fmt.Fprintln(os.Stderr, serverPayment.Serve(liPayment))
	}()

	go func() {
		fmt.Fprintln(os.Stderr, serverShipment.Serve(liShipment))
	}()

	return nil
}
