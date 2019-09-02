package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	delay time.Duration
	mu    sync.RWMutex

	allowedIPs []net.IP

	mux *http.ServeMux
}

type Adapter func(http.Handler) http.Handler

type errorRes struct {
	Error string `json:"error"`
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

func (s *Server) withDelay() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-time.After(s.GetDelay())
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) withIPRestriction() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(s.allowedIPs) != 0 {
				ip, err := userIP(r)
				if err != nil {
					log.Print(err)
					return
				}

				passed := false
				for _, aIP := range s.allowedIPs {
					if ip.Equal(aIP) {
						passed = true
						break
					}
				}

				if !passed {
					b, _ := json.Marshal(errorRes{Error: "IP address is not allowed"})

					w.WriteHeader(http.StatusForbidden)
					w.Write(b)

					return
				}
			}

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

func userIP(r *http.Request) (net.IP, error) {
	tcIP := r.Header.Get("True-Client-IP")

	// 未検証で信じる
	// DO NOT COPY the following code
	if tcIP != "" {
		userIP := net.ParseIP(tcIP)
		if userIP == nil {
			return nil, fmt.Errorf("userip: %q is not IP:port", tcIP)
		}
		return userIP, nil
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}
	return userIP, nil
}

func RunServer(paymentPort, shipmentPort int, dataDir string, allowedIPs []net.IP) (*ServerPayment, *ServerShipment, error) {
	liPayment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: paymentPort})
	if err != nil {
		return nil, nil, err
	}

	liShipment, err := net.ListenTCP("tcp", &net.TCPAddr{Port: shipmentPort})
	if err != nil {
		return nil, nil, err
	}

	pay := NewPayment(allowedIPs)
	serverPayment := &http.Server{
		Handler: pay,
	}

	ship := NewShipment(false, dataDir, allowedIPs)
	serverShipment := &http.Server{
		Handler: ship,
	}

	go func() {
		log.Print(serverPayment.Serve(liPayment))
	}()

	go func() {
		log.Print(serverShipment.Serve(liShipment))
	}()

	return pay, ship, nil
}
