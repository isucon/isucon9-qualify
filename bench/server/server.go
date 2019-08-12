package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/external/payment"
	"github.com/isucon/isucon9-qualify/external/shipment"
)

type Server struct {
	delay time.Duration
	mu    sync.RWMutex

	mux *http.ServeMux
}

type Adapter func(http.Handler) http.Handler

func NewShipment() *Server {
	s := &Server{}

	s.mux = http.NewServeMux()

	s.mux.Handle("/create", apply(http.HandlerFunc(shipment.CreateHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/request", apply(http.HandlerFunc(shipment.RequestHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/accept", apply(http.HandlerFunc(shipment.AcceptHandler), s.withDelay(), s.withIPRestriction()))
	s.mux.Handle("/status", apply(http.HandlerFunc(shipment.StatusHandler), s.withDelay(), s.withIPRestriction()))

	return s
}

func NewPayment() *Server {
	s := &Server{}

	s.mux = http.NewServeMux()

	// cardだけはdelayなし
	s.mux.Handle("/card", apply(http.HandlerFunc(payment.CardHandler), s.withIPRestriction()))
	s.mux.Handle("/token", apply(http.HandlerFunc(payment.TokenHandler), s.withDelay(), s.withIPRestriction()))

	return s
}

func (s *Server) SetDelay(d time.Duration) {
	s.mu.Lock()
	s.delay = d
	s.mu.Unlock()
}

func (s *Server) GetDelay() time.Duration {
	s.mu.RLock()
	d := s.delay
	s.mu.RUnlock()
	return d
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func tmp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withDelay() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			<-time.After(s.GetDelay())
		})
	}
}

func (s *Server) withIPRestriction() Adapter {
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
