package bitrise

import (
	"net/http"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
)

// RegisterFinishParams ...
type RegisterFinishParams struct {
	Config           string            `json:"config,omitempty"`
	Envs             map[string]string `json:"envs,omitempty"`
	Mode             string            `json:"mode,omitempty"`
	OrganizationSlug string            `json:"organization_slug,omitempty"`
	ProjectType      string            `json:"project_type,omitempty"`
	StackID          string            `json:"stack_id,omitempty"`
}

// RegisterFinish ...
func (c *Client) RegisterFinish(appSlug, bitriseYML string, envs map[string]string, mode, organizationSlug, projectType, stackID string) (*http.Response, error) {
	p := RegisterFinishParams{
		Config:           bitriseYML,
		Envs:             envs,
		Mode:             mode,
		OrganizationSlug: organizationSlug,
		ProjectType:      projectType,
		StackID:          stackID,
	}

	req, err := c.newRequest(http.MethodPost, appSlug+"/finish", p)
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
