package bitrise

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
	return fmt.Sprintf("apps/%s/register-ssh-key", appSlug)
}

// RegisterSSHKey ...
func (c *Client) RegisterSSHKey(appSlug string, params RegisterSSHKeyParams) error {
	req, err := c.newRequest(http.MethodPost, RegisterSSHKeyURL(appSlug), params)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}
