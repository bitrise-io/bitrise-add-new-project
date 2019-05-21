package bitrise

import (
	"fmt"
	"net/http"
)

// RegisterWebhookURL ...
func RegisterWebhookURL(appSlug string) string {
	return fmt.Sprintf("apps/%s/register-webhook", appSlug)
}

// RegisterWebhook ...
func (c *Client) RegisterWebhook(appSlug string) error {
	req, err := c.newRequest(http.MethodPost, RegisterWebhookURL(appSlug), nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}
