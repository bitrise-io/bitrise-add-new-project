package bitriseio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
	"github.com/bitrise-io/go-utils/log"
)

const apiVersion = "v0.1"

var baseURL = "https://api.bitrise.io/" + apiVersion + "/"

// Client ...
type Client struct {
	BaseURL *url.URL

	client *http.Client
	token  string

	Apps *AppsService
}

// NewClient ...
func NewClient(token string) (*Client, error) {
	httpClient := http.DefaultClient
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		BaseURL: baseURL,
		client:  httpClient,
		token:   token,
	}
	c.Apps = &AppsService{
		client: c,
	}

	return c, nil
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

func (c *Client) do(req *http.Request, v interface{}) error {
	log.Debugf("Request:")
	if err := httputil.PrintRequest(req); err != nil {
		log.Debugf("Failed to print request: %s", err)
	}

	resp, err := c.client.Do(req)

	log.Debugf("Response:")
	if err := httputil.PrintResponse(resp); err != nil {
		log.Debugf("Failed to print response: %s", err)
	}

	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Debugf("Failed to close response body: %s", err)
		}
	}()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}
	err := &ErrorResponse{Response: r}
	if err := json.NewDecoder(r.Body).Decode(&err); err != nil && err != io.EOF {
		return err
	}
	return err
}

// ErrorResponse ...
type ErrorResponse struct {
	Status int    `json:"status"`
	Err    string `json:"error"`

	ErrTypeCode int    `json:"error_type_code"`
	ErrMsg      string `json:"error_msg"`

	Message string `json:"message"`

	Response *http.Response
}

// Error ...
func (r ErrorResponse) Error() string {
	m := fmt.Sprintf("%v %v: %d\n", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode)

	switch {
	case len(r.Err) > 0:
		if r.Status != 0 {
			m += fmt.Sprintf("%d: ", r.Status)
		}
		m += r.Err + "\n"
	case len(r.ErrMsg) > 0:
		if r.ErrTypeCode != 0 {
			m += fmt.Sprintf("%d: ", r.ErrTypeCode)
		}
		m += r.ErrMsg + "\n"
	case len(r.Message) > 0:
		m += r.Message + "\n"
	}

	return m
}
