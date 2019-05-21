package bitrise

import (
	"fmt"
	"net/http"
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

// RegisterFinishResponse ...
type RegisterFinishResponse struct {
	Status string `json:"status,omitempty"`
	Slug   string `json:"slug,omitempty"`
}

// RegisterFinishURL ...
func RegisterFinishURL(appSlug string) string {
	return fmt.Sprintf("apps/%s/finish", appSlug)
}

// RegisterFinish ...
func (c *Client) RegisterFinish(appSlug string, params RegisterFinishParams) (string, error) {
	req, err := c.newRequest(http.MethodPost, RegisterFinishURL(appSlug), params)
	if err != nil {
		return "", err
	}
	var resp RegisterFinishResponse
	if err := c.do(req, &resp); err != nil {
		return "", err
	}
	return resp.Slug, nil
}
