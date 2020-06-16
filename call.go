package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Client is a HTTP client
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// newRequest creates a HTTP request
func (c *Client) newRequest(method, path string, body []byte) (*http.Request, error) {

	p, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL.ResolveReference(p)

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-ExperimentalApi", "opt-in")

	cerr := errors.New("missing credentials")
	admin, ok := os.LookupEnv("ADMIN_USER")
	if !ok {
		return nil, cerr
	}
	password, ok := os.LookupEnv("ADMIN_PASS")
	if !ok {
		return nil, cerr
	}
	req.SetBasicAuth(admin, password)

	return req, nil
}

// do makes a HTTP request
func (c *Client) do(req *http.Request) (*http.Response, error) {

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, err
}

// callJSD makes the outbound request
func callJSD(m *Message) error {

	jsdURL, err := url.Parse(os.Getenv("JSD_URL"))
	if err != nil {
		return err
	}

	c := &Client{
		BaseURL:    jsdURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	req, err := c.newRequest(m.Method, m.URI, m.Payload)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}

	log.Printf("JSD response code: %v", resp.StatusCode)

	return nil
}
