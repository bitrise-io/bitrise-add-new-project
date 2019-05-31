package bitriseio

import (
	"net/http"
)

// RegisterURL ...
const RegisterURL = AppsServiceURL + "register"

// RegisterParams ...
type RegisterParams struct {
	GitOwner    string `json:"git_owner"`
	GitRepoSlug string `json:"git_repo_slug"`
	IsPublic    bool   `json:"is_public"`
	Provider    string `json:"provider"`
	RepoURL     string `json:"repo_url"`
}

// Register ...
func (s *AppsService) Register(params RegisterParams) (*AppService, error) {
	type Params struct {
		RegisterParams
		Type string `json:"type"`
	}
	p := Params{RegisterParams: params}
	p.Type = "git"

	req, err := s.client.newRequest(http.MethodPost, RegisterURL, p)
	if err != nil {
		return nil, err
	}
	type RegisterResponse struct {
		Slug string `json:"slug"`
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
