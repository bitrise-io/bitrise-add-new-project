package bitrise

import (
	"net/http"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
)

// RegisterWebhook ...
func (c *Client) RegisterWebhook(appSlug string) (*http.Response, error) {
	req, err := c.newRequest(http.MethodPost, appSlug+"/register-webhook", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req, nil)
	httputil.PrintResponse(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
