package bitrise

import (
	"net/http"
)

// RegisterURL ...
const RegisterURL = "apps/register"

// RegisterParams ...
type RegisterParams struct {
	GitOwner    string `json:"git_owner,omitempty"`
	GitRepoSlug string `json:"git_repo_slug,omitempty"`
	IsPublic    bool   `json:"is_public,omitempty"`
	Provider    string `json:"provider,omitempty"`
	RepoURL     string `json:"repo_url,omitempty"`
	Type        string `json:"type,omitempty"`
}

// RegisterResponse ...
type RegisterResponse struct {
	Status string `json:"status,omitempty"`
	Slug   string `json:"slug,omitempty"`
}

// Register ...
func (c *Client) Register(params RegisterParams) (string, error) {
	req, err := c.newRequest(http.MethodPost, RegisterURL, params)
	if err != nil {
		return "", err
	}
	var resp RegisterResponse
	if err := c.do(req, &resp); err != nil {
		return "", err
	}
	return resp.Slug, nil
}
