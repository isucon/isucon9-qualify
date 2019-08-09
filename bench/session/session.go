package session

import (
	"fmt"
	"io"
	"io/ioutil"
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
	CSRFToken  string
	httpClient *http.Client
}

type TargetURLs struct {
	AppURL      *url.URL
	PaymentURL  *url.URL
	ShipmentURL *url.URL
}

var (
	ShareTargetURLs *TargetURLs
)

func SetShareTargetURLs(appURL, paymentURL, shipmentURL string) error {
	var err error
	ShareTargetURLs, err = newTargetURLs(appURL, paymentURL, shipmentURL)
	if err != nil {
		return err
	}

	return nil
}

func newTargetURLs(appURL, paymentURL, shipmentURL string) (*TargetURLs, error) {
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
	if len(spath) > 0 {
		u.Path = spath
	}

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

func checkStatusCode(res *http.Response, expectedStatusCode int) (msg string, err error) {
	if res.StatusCode != expectedStatusCode {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "bodyの読み込みに失敗しました", err
		}
		return fmt.Sprintf("got response status code %d; expected %d", res.StatusCode, expectedStatusCode), fmt.Errorf("status code: %d; body: %s", res.StatusCode, b)
	}

	return "", nil
}

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
