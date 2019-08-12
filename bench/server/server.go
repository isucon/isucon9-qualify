package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/external/payment"
)

type Server struct {
	delay time.Duration
	mu    sync.RWMutex

	mux *http.ServeMux
}

type Adapter func(http.Handler) http.Handler

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

	pay := NewPayment()
	serverPayment := &http.Server{
		Handler: pay,
	}

	ship := NewShipment()
	serverShipment := &http.Server{
		Handler: ship,
	}

	go func() {
		fmt.Fprintln(os.Stderr, serverPayment.Serve(liPayment))
	}()

	go func() {
		fmt.Fprintln(os.Stderr, serverShipment.Serve(liShipment))
	}()

	return nil
}
