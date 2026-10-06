package bitriseio

import (
	"fmt"
)

// RegisterWebhookURL ...
func RegisterWebhookURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/register-webhook", appSlug)
}

// RegisterWebhook ...
func (s *AppService) RegisterWebhook() error {
	req, err := s.client.newPostRequest(RegisterWebhookURL(s.Slug), nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
