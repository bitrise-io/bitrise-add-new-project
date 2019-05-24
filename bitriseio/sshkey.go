package bitriseio

import (
	"fmt"
	"net/http"
)

// RegisterSSHKeyParams ...
type RegisterSSHKeyParams struct {
	AuthSSHPrivateKey                string `json:"auth_ssh_private_key,omitempty"`
	AuthSSHPublicKey                 string `json:"auth_ssh_public_key,omitempty"`
	IsRegisterKeyIntoProviderService bool   `json:"is_register_key_into_provider_service,omitempty"`
}

// RegisterSSHKeyURL ...
func RegisterSSHKeyURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/register-ssh-key", appSlug)
}

// RegisterSSHKey ...
func (s *AppsService) RegisterSSHKey(appSlug string, params RegisterSSHKeyParams) error {
	req, err := s.client.newRequest(http.MethodPost, RegisterSSHKeyURL(appSlug), params)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
