package bitriseio

import (
	"fmt"
	"net/http"
)

// RegisterFinishParams ...
type RegisterFinishParams struct {
	Config           string            `json:"config,omitempty"` // in case of local run we do not have scan result, so our only option is Mode: manual + Config: any default config id
	Envs             map[string]string `json:"envs,omitempty"`
	Mode             string            `json:"mode,omitempty"`
	OrganizationSlug string            `json:"organization_slug,omitempty"` // leave it empty if the user will be the owne
	ProjectType      string            `json:"project_type,omitempty"`
	StackID          string            `json:"stack_id,omitempty"`
}

// RegisterFinishResponse ...
type RegisterFinishResponse struct {
	Status                    string `json:"status,omitempty"`
	BuildTriggerToken         string `json:"build_trigger_token,omitempty"`
	BranchName                string `json:"branch_name,omitempty"`
	IsWebhookAutoRegSupported bool   `json:"is_webhook_auto_reg_supported,omitempty"`
}

// RegisterFinishURL ...
func RegisterFinishURL(appSlug string) string {
	return fmt.Sprintf(AppsServiceURL+"%s/finish", appSlug)
}

// RegisterFinish ...
func (s *AppService) RegisterFinish(params RegisterFinishParams) (*RegisterFinishResponse, error) {
	params.Mode = "manual"
	params.Config = "default-ios-config"

	req, err := s.client.newRequest(http.MethodPost, RegisterFinishURL(s.Slug), params)
	if err != nil {
		return nil, err
	}
	var resp RegisterFinishResponse
	if err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
