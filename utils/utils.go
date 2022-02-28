package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36"
	localCertFile    = "test.crt"
)

type Client struct {
	httpClient *http.Client
	BaseURL    *url.URL
	UserAgent  string
}

func NewClient(httpClient *http.Client, apiURL string) *Client {

	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadFile(localCertFile)
	if err != nil {
		log.Fatalf("Failed to append %q to RootCAs: %v", localCertFile, err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("No certs appended, using system certs only")
	}

	config := &tls.Config{
		RootCAs: rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: tr,
		}
	}

	baseURL, _ := url.Parse(apiURL)

	c := &Client{
		httpClient: httpClient,
		UserAgent:  defaultUserAgent,
		BaseURL:    baseURL,
	}
	return c
}

// Do sends an API request and returns the API response. The API response is JSON
// decoded and stored in the value pointed to by v.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.httpClient.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil && err != io.EOF {
		return nil, err
	}
	return resp, nil
}

// NewRequest creates an API request. The given URL is relative to the Client's
// BaseURL.
func (c *Client) NewRequest(method, url string, body interface{}) (*http.Request, error) {

	u, err := c.BaseURL.Parse(url)
	if err != nil {
		return nil, err
	}
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func checkResponse(r *http.Response) error {
	status := r.StatusCode
	if status >= 200 && status <= 299 {
		return nil
	}

	return fmt.Errorf("request failed with status %d", status)
}
