package bitrise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/godrei/go-appstoreconnect/debug"
)

const (
	defaultBaseURL = "https://api.bitrise.io/v0.1/apps/"
)

// Client ...
type Client struct {
	BaseURL *url.URL

	client *http.Client
	token  string
}

// NewClient ...
func NewClient(token string) (*Client, error) {
	httpClient := http.DefaultClient
	baseURL, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		BaseURL: baseURL,
		client:  httpClient,
		token:   token,
	}, nil
}

// newRequest creates a new http.Request
func (c *Client) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
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

	req.Header.Set("Authorization", c.token)

	return req, nil
}

// ErrorResponseError ...
type ErrorResponseError struct {
	Code   string      `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	ID     string      `json:"id,omitempty"`
	Title  string      `json:"title,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Source interface{} `json:"source,omitempty"`
}

// ErrorResponse ...
type ErrorResponse struct {
	Response *http.Response
	Errors   []ErrorResponseError `json:"errors,omitempty"`
}

// Error ...
func (r ErrorResponse) Error() string {
	m := fmt.Sprintf("%v %v: %d\n", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode)
	var s string

	for _, err := range r.Errors {
		m += s + fmt.Sprintf("- %v %v", err.Title, err.Detail)
		s = "\n"
	}

	return m
}

func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	log.Debugf("Request:")
	httputil.PrintRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("Response:")
	debug.PrintResponse(resp)

	if err := checkResponse(resp); err != nil {
		return resp, err
	}

	if v != nil {
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}

	return resp, err
}
