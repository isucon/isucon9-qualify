package session

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/xerrors"
)

const (
	DefaultAPITimeout = 10

	userAgent = "benchmarker/isucon9-qualify"
)

type Session struct {
	httpClient *http.Client
}

type TargetURLs struct {
	AppURL      *url.URL
	PaymentURL  *url.URL
	ShipmentURL *url.URL
}

var (
	ShareTargetURLs TargetURLs
)

func NewTargetURLs(appURL, paymentURL, shipmentURL string) (*TargetURLs, error) {
	if len(appURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	if len(paymentURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	if len(shipmentURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	appParsedURL, err := url.ParseRequestURI(appURL)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse url: %s: %w", appURL, err)
	}

	paymentParsedURL, err := url.ParseRequestURI(paymentURL)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse url: %s: %w", paymentURL, err)
	}

	shipmentParsedURL, err := url.ParseRequestURI(shipmentURL)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse url: %s: %w", shipmentURL, err)
	}

	return &TargetURLs{
		AppURL:      appParsedURL,
		PaymentURL:  paymentParsedURL,
		ShipmentURL: shipmentParsedURL,
	}, nil
}

func NewSession() (*Session, error) {
	jar, _ := cookiejar.New(&cookiejar.Options{})

	s := &Session{
		httpClient: &http.Client{
			Transport: &http.Transport{},
			Jar:       jar,
			Timeout:   time.Duration(DefaultAPITimeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return fmt.Errorf("redirect attempted")
			},
		},
	}

	return s, nil
}

func (s *Session) newGetRequest(u *url.URL, spath string) (*http.Request, error) {
	u.Path = spath

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func (s *Session) newPostRequest(u *url.URL, spath, contentType string, body io.Reader) (*http.Request, error) {
	u.Path = spath

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
