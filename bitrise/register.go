package bitrise

import (
	"net/http"

	"github.com/bitrise-io/bitrise-add-new-project/httputil"
)

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
func (c *Client) Register(owner, repoSlug string, public bool, provider, repoURL, repoType string) (*http.Response, error) {
	p := RegisterParams{
		IsPublic:    public,
		GitOwner:    owner,
		GitRepoSlug: repoSlug,
		Provider:    provider,
		RepoURL:     repoURL,
		Type:        repoType,
	}

	req, err := c.newRequest(http.MethodPost, "register", p)
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
