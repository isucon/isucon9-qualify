package server

import (
	"net/http"

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
