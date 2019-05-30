package bitriseio

import (
	"net/http"
)

// RegisterURL ...
const RegisterURL = AppsServiceURL + "register"

// RegisterParams ...
type RegisterParams struct {
	GitOwner    string `json:"git_owner,omitempty"`
	GitRepoSlug string `json:"git_repo_slug,omitempty"`
	IsPublic    bool   `json:"is_public,omitempty"`
	Provider    string `json:"provider,omitempty"`
	RepoURL     string `json:"repo_url,omitempty"`
	Type        string `json:"type,omitempty"`
}

// Register ...
func (s *AppsService) Register(params RegisterParams) (*AppService, error) {
	req, err := s.client.newRequest(http.MethodPost, RegisterURL, params)
	if err != nil {
		return nil, err
	}
	type RegisterResponse struct {
		Slug string `json:"slug,omitempty"`
	}
	var resp RegisterResponse
	if err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &AppService{
		client: s.client,
		Slug:   resp.Slug,
	}, nil
}
