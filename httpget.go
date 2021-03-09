package httpget

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// Client for the http requets
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	UserAgent  string
}

// NewClient returns a new Client
func NewClient(httpClient *http.Client, baseU string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(baseU)
	jar, _ := cookiejar.New(nil)
	httpClient.Jar = jar
	c := &Client{HTTPClient: httpClient, BaseURL: baseURL}
	return c
}

// NewRequest creates a HTTP request
func (c *Client) NewRequest(method, urlStr string, json bool, headers map[string]string, body []byte) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if json {
		req.Header.Set("Content-Type", "application/json")
	}

	if headers != nil {
		for k, v := range headers {

			req.Header.Set(k, v)
		}
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil

}

// Do executes the HTTP request
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}
	req.WithContext(ctx)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}
	return resp, err
}
