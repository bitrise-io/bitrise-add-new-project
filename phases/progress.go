package phases

import "github.com/bitrise-io/bitrise/models"

// Progress ...
type Progress struct {
	OrganizationSlug string
	Public           bool

	RepoURL RepoDetails

	SSHPrivateKey  string
	SSHPublicKey   string
	RegisterSSHKey bool

	BitriseYML      models.BitriseDataModel
	PrimaryWorkflow string
	ProjectType     string

	Stack string

	AddWebhook bool

	Codesign CodesignResult
}
