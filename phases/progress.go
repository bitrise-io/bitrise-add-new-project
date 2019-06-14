package phases

import (
	"github.com/bitrise-io/bitrise-add-new-project/sshutil"
	"github.com/bitrise-io/bitrise/models"
)

// Progress ...
type Progress struct {
	OrganizationSlug string
	Public           bool

	RepoDetails RepoDetails

	SSHKeys        sshutil.SSHKeyPair
	RegisterSSHKey bool

	BitriseYML      models.BitriseDataModel
	PrimaryWorkflow string
	ProjectType     string

	Stack string

	AddWebhook bool

	Codesign CodesignResult
}
