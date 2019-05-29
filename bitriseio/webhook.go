package bitriseio

import (
	"fmt"
	"net/http"
)

// RegisterWebhookURL ...
func RegisterWebhookURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/register-webhook", appSlug)
}

// RegisterWebhook ...
func (s *AppService) RegisterWebhook() error {
	req, err := s.client.newRequest(http.MethodPost, RegisterWebhookURL(s.Slug), nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
