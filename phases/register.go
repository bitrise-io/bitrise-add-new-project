package phases

import (
	"fmt"

	"github.com/bitrise-io/go-utils/fileutil"

	"github.com/bitrise-io/bitrise-add-new-project/bitriseio"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/xcode-project/pretty"
	"gopkg.in/yaml.v2"
)

// CreateProjectParams ...
type CreateProjectParams struct {
	Repository      bitriseio.RegisterParams
	SSHKey          bitriseio.RegisterSSHKeyParams
	RegisterWebhook bool
	Project         bitriseio.RegisterFinishParams
	BitriseYML      bitriseio.BitriseYMLParams
	TriggerBuild    bitriseio.TriggerBuildParams
	Keystore        bitriseio.UploadKeystoreParams
	KeystorePth     string
}

func toRegistrationParams(progress Progress) (*CreateProjectParams, error) {
	bitriseYML, err := yaml.Marshal(progress.BitriseYML)
	if err != nil {
		return nil, err
	}
	bitriseYMLstr := string(bitriseYML)

	privateKey, err := fileutil.ReadStringFromFile(progress.SSHPrivateKeyPth)
	if err != nil {
		return nil, err
	}

	var publicKey string
	if progress.RegisterSSHKey {
		var err error
		publicKey, err = fileutil.ReadStringFromFile(progress.SSHPublicKeyPth)
		if err != nil {
			return nil, err
		}
	}

	params := CreateProjectParams{}
	params.Repository = bitriseio.RegisterParams{
		GitOwner:    progress.RepoOwner,
		GitRepoSlug: progress.RepoSlug,
		IsPublic:    progress.Public,
		Provider:    progress.RepoProvider,
		RepoURL:     progress.RepoURL,
		Type:        progress.RepoType,
	}
	params.RegisterWebhook = progress.AddWebhook
	params.SSHKey = bitriseio.RegisterSSHKeyParams{
		AuthSSHPrivateKey:                privateKey,
		AuthSSHPublicKey:                 publicKey,
		IsRegisterKeyIntoProviderService: progress.RegisterSSHKey,
	}
	params.Project = bitriseio.RegisterFinishParams{
		Config:           bitriseYMLstr,
		Envs:             nil,
		Mode:             "manual",
		OrganizationSlug: progress.Account,
		ProjectType:      "",
		StackID:          progress.Stack,
	}
	params.KeystorePth = progress.Codesign.KeystorePath
	params.Keystore = bitriseio.UploadKeystoreParams{
		Password:    progress.Codesign.Password,
		Alias:       progress.Codesign.Alias,
		KeyPassword: progress.Codesign.KeyPassword,
	}
	return &params, nil
}

// Register ...
func Register(token string, progress Progress) error {
	log.Infof("Register")

	params, err := toRegistrationParams(progress)
	if err != nil {
		return err
	}

	log.Printf("Provided params: %s", pretty.Object(params))

	client, err := bitriseio.NewClient(token)
	if err != nil {
		return err
	}
	service := client.Apps
	slug, err := service.Register(params.Repository)
	if err != nil {
		return err
	}
	if !params.Repository.IsPublic {
		if err := service.RegisterSSHKey(slug, params.SSHKey); err != nil {
			return err
		}
	}
	if params.RegisterWebhook {
		if err := service.RegisterWebhook(slug); err != nil {
			return err
		}
	}
	resp, err := service.RegisterFinish(slug, params.Project)
	if err != nil {
		return err
	}

	fmt.Println(pretty.Object(resp))

	if err := service.BitriseYML(slug, params.BitriseYML); err != nil {
		return err
	}

	if params.KeystorePth != "" {
		if err := service.UploadKeystore(slug, params.KeystorePth, params.Keystore); err != nil {
			return err
		}
	}

	if err := service.TriggerBuild(slug, params.TriggerBuild); err != nil {
		return err
	}

	log.Donef("Project created: https://app.bitriseio.io/app/%s", slug)
	return nil
}
