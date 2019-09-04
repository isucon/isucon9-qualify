package session

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/morikuni/failure"
)

const (
	DefaultAPITimeout = 10

	userAgent = "benchmarker/isucon9-qualify"
)

type Session struct {
	UserID     int64
	csrfToken  string
	httpClient *http.Client
}

type TargetURLs struct {
	AppURL      url.URL
	TargetHost  string
	PaymentURL  url.URL
	ShipmentURL url.URL
}

var (
	ShareTargetURLs *TargetURLs
)

func SetShareTargetURLs(appURL, targetHost, paymentURL, shipmentURL string) error {
	var err error
	ShareTargetURLs, err = newTargetURLs(appURL, targetHost, paymentURL, shipmentURL)
	if err != nil {
		return err
	}

	return nil
}

func newTargetURLs(appURL, targetHost, paymentURL, shipmentURL string) (*TargetURLs, error) {
	if len(appURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	if len(paymentURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	if len(shipmentURL) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	appParsedURL, err := urlParse(appURL)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("failed to parse url: %s", appURL))
	}

	paymentParsedURL, err := urlParse(paymentURL)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("failed to parse url: %s", paymentURL))
	}

	shipmentParsedURL, err := urlParse(shipmentURL)
	if err != nil {
		return nil, failure.Wrap(err, failure.Messagef("failed to parse url: %s", shipmentURL))
	}

	return &TargetURLs{
		AppURL:      *appParsedURL,
		TargetHost:  targetHost,
		PaymentURL:  *paymentParsedURL,
		ShipmentURL: *shipmentParsedURL,
	}, nil
}

func urlParse(ref string) (*url.URL, error) {
	u, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}

	if u.Host == "" {
		return nil, fmt.Errorf("host is empty")
	}

	return &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
	}, nil
}

func (s *Session) newGetRequest(u url.URL, spath string) (*http.Request, error) {
	if len(spath) > 0 {
		u.Path = spath
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Host = ShareTargetURLs.TargetHost
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func (s *Session) newGetRequestWithQuery(u url.URL, spath string, q url.Values) (*http.Request, error) {
	if len(spath) > 0 {
		u.Path = spath
	}

	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Host = ShareTargetURLs.TargetHost
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func (s *Session) newPostRequest(u url.URL, spath, contentType string, body io.Reader) (*http.Request, error) {
	u.Path = spath

	req, err := http.NewRequest(http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Host = ShareTargetURLs.TargetHost
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

func checkStatusCode(res *http.Response, expectedStatusCode int) error {
	prefixMsg := fmt.Sprintf("%s %s", res.Request.Method, res.Request.URL.Path)

	if res.StatusCode != expectedStatusCode {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return failure.Wrap(err, failure.Message(prefixMsg+": bodyの読み込みに失敗しました"))
		}
		return failure.Translate(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), fails.ErrApplication,
			failure.Messagef("%s: got response status code %d; expected %d", prefixMsg, res.StatusCode, expectedStatusCode),
		)
	}

	return nil
}

func checkStatusCodeWithMsg(res *http.Response, expectedStatusCode int, msg string) error {
	prefixMsg := fmt.Sprintf("%s %s", res.Request.Method, res.Request.URL.Path)

	if res.StatusCode != expectedStatusCode {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return failure.Wrap(err, failure.Message(prefixMsg+": bodyの読み込みに失敗しました "+msg))
		}
		return failure.Translate(fmt.Errorf("status code: %d; body: %s", res.StatusCode, b), fails.ErrApplication,
			failure.Messagef("%s: got response status code %d; expected %d %s", prefixMsg, res.StatusCode, expectedStatusCode, msg),
		)
	}

	return nil
}

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	res, err := s.httpClient.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok {
			if nerr.Timeout() {
				return nil, failure.Translate(err, fails.ErrTimeout)
			} else if nerr.Temporary() {
				return nil, failure.Translate(err, fails.ErrTemporary)
			}
		}

		return nil, err
	}

	return res, nil
}
