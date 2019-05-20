package bitrise

import (
	"net/http"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
)

// RegisterSSHKeyParams ...
type RegisterSSHKeyParams struct {
	PrivateKey string `json:"auth_ssh_private_key,omitempty"`
	PublicKey  string `json:"auth_ssh_public_key,omitempty"`
	IsRegister bool   `json:"is_register_key_into_provider_service,omitempty"`
}

// RegisterSSHKey ...
func (c *Client) RegisterSSHKey(appSlug, privateKey, publicKey string, isRegister bool) (*http.Response, error) {
	p := RegisterSSHKeyParams{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		IsRegister: isRegister,
	}

	req, err := c.newRequest(http.MethodPost, appSlug+"/register-ssh-key", p)
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
