package bitriseio

import (
	"fmt"
	"net/http"

	"github.com/bitrise-io/bitrise-add-new-project/sshutil"
	"github.com/bitrise-io/go-utils/log"
)

// RegisterSSHKeyParams ...
type RegisterSSHKeyParams struct {
	AuthSSHPrivateKey                string `json:"auth_ssh_private_key"`
	AuthSSHPublicKey                 string `json:"auth_ssh_public_key,omitempty"`
	IsRegisterKeyIntoProviderService bool   `json:"is_register_key_into_provider_service"`
	Username                         string
}

// RegisterSSHKeyURL ...
func RegisterSSHKeyURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/register-ssh-key", appSlug)
}

func (s *AppService) registerSSHKeyRequest(params RegisterSSHKeyParams) error {
	req, err := s.client.newRequest(http.MethodPost, RegisterSSHKeyURL(s.Slug), params)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

// RegisterSSHKey ...
func (s *AppService) RegisterSSHKey(params RegisterSSHKeyParams, repoURL string) error {
	if err := s.registerSSHKeyRequest(params); err != nil {
		if !params.IsRegisterKeyIntoProviderService {
			return err
		}

		log.Errorf("Failed to automatically register SSH key. Falling back to manual registration.")
		params.IsRegisterKeyIntoProviderService = false

		if err := sshutil.ValidateSSHAddedManually(sshutil.SSHRepo{
			Keys: sshutil.SSHKeyPair{
				PublicKey:  []byte(params.AuthSSHPublicKey),
				PrivateKey: []byte(params.AuthSSHPrivateKey),
			},
			URL:        repoURL,
			Username:   params.Username,
		}); err != nil {
			return err
		}

		return s.registerSSHKeyRequest(params)
	}
	return nil
}
