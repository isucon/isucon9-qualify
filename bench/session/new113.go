// +build go1.13

package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func NewSession() (*Session, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{})

	s := &Session{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// HTTPの時は無視されるだけ
					ServerName: ShareTargetURLs.TargetHost,
				},
				// TLSClientConfigを上書きしてもHTTP/2を使えるように
				ForceAttemptHTTP2: true,
			},
			Jar:     jar,
			Timeout: time.Duration(DefaultAPITimeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return fmt.Errorf("redirect attempted")
			},
		},
	}

	return s, nil
}

func NewSessionForInialize() (*Session, error) {
	s := &Session{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// HTTPの時は無視されるだけ
					ServerName: ShareTargetURLs.TargetHost,
				},
				// TLSClientConfigを上書きしてもHTTP/2を使えるように
				ForceAttemptHTTP2: true,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return fmt.Errorf("redirect attempted")
			},
		},
	}

	return s, nil
}
