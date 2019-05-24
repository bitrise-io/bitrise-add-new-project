package phases

import "github.com/bitrise-io/bitrise/models"

// Progress ...
type Progress struct {
	Account      *string
	Public       *bool
	Repo         *string
	RepoURL      *string
	RepoProvider *string
	RepoOwner    *string
	RepoSlug     *string
	RepoType     *string

	SSHPrivateKeyPth string
	SSHPublicKeyPth  string
	RegisterSSHKey   bool

	BitriseYML      models.BitriseDataModel
	PrimaryWorkflow string
	ProjectType     string

	Stack *string

	AddWebhook   *bool
	AutoCodesign *bool
}
