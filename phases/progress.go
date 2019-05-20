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
	PrivateKey   *string
	BitriseYML   *models.BitriseDataModel
	Stack        *string
	AddWebhook   *bool
	AutoCodesign *bool
}
