package phases

import (
	"encoding/json"
	"os"
)

// Progress ...
type Progress struct {
	filePath     string
	Account      *string `json:"account,omitempty"`
	Public       *bool   `json:"public,omitempty"`
	Repo         *string `json:"repo,omitempty"`
	RepoURL      *string `json:"repo_url,omitempty`
	RepoProvider *string `json:"repo_provider,omitempty`
	RepoOwner    *string `json:"repo_owner,omitempty`
	RepoSlug     *string `json:"repo_slug,omitempty`
	RepoType     *string `json:"repo_type,omitempty`
	PrivateKey   *string `json:"private_key,omitempty"`
	BitriseYML   *string `json:"bitrise_yml,omitempty"`
	Stack        *string `json:"stack,omitempty"`
	AddWebhook   *bool   `json:"webhook,omitempty"`
	AutoCodesign *bool   `json:"codesign,omitempty"`
}

// Store ...
func (p *Progress) Store() error {
	f, err := os.Create(p.filePath)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(p)
}

// Destroy ...
func (p *Progress) Destroy() error {
	if err := os.RemoveAll(p.filePath); err != nil {
		return err
	}
	return nil
}

// LoadProgress ...
func LoadProgress(filePath string) (*Progress, error) {
	switch f, err := os.Open(filePath); {
	case os.IsNotExist(err):
		return &Progress{filePath: filePath}, nil
	case err != nil:
		return nil, err
	default:
		var p Progress
		if err := json.NewDecoder(f).Decode(&p); err != nil {
			return nil, err
		}
		p.filePath = filePath
		return &p, nil
	}
}
