package phases

import (
	"bufio"
	"fmt"
	"io"

	"github.com/bitrise-io/bitrise-add-new-project/bitriseio"
	"github.com/bitrise-io/bitrise-add-new-project/httputil"
	codesigndocBitriseio "github.com/bitrise-io/codesigndoc/bitriseio"
	"github.com/bitrise-io/codesigndoc/bitriseio/bitrise"
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
	BitriseYML      string
	WorkflowID      string
	Branch          string
	Keystore        bitriseio.UploadKeystoreParams
	KeystorePth     string
	CodesignIOS     CodesignResultsIOS
}

func toRegistrationParams(progress Progress) (*CreateProjectParams, error) {
	bitriseYML, err := yaml.Marshal(progress.BitriseYML)
	if err != nil {
		return nil, fmt.Errorf("bitrise.yml marshal failed: %s", err)
	}
	bitriseYMLstr := string(bitriseYML)

	params := CreateProjectParams{}
	params.Repository = bitriseio.RegisterParams{
		GitOwner:    progress.RepoDetails.Owner,
		GitRepoSlug: progress.RepoDetails.Slug,
		IsPublic:    progress.Public,
		Provider:    progress.RepoDetails.Provider,
		RepoURL:     progress.RepoDetails.URL,
	}
	params.RegisterWebhook = progress.AddWebhook

	params.SSHKey = bitriseio.RegisterSSHKeyParams{
		AuthSSHPrivateKey:                string(progress.SSHKeys.PrivateKey),
		AuthSSHPublicKey:                 string(progress.SSHKeys.PublicKey),
		IsRegisterKeyIntoProviderService: progress.RegisterSSHKey,
		Username:                         progress.RepoDetails.SSHUsername,
	}

	params.Project = bitriseio.RegisterFinishParams{
		OrganizationSlug: progress.OrganizationSlug,
		ProjectType:      progress.ProjectType,
		StackID:          progress.Stack,
	}
	params.CodesignIOS = progress.Codesign.IOS
	params.KeystorePth = progress.Codesign.Android.KeystorePath
	params.Keystore = bitriseio.UploadKeystoreParams{
		Password:    progress.Codesign.Android.Password,
		Alias:       progress.Codesign.Android.Alias,
		KeyPassword: progress.Codesign.Android.KeyPassword,
	}
	params.BitriseYML = bitriseYMLstr
	params.WorkflowID = progress.PrimaryWorkflow
	params.Branch = progress.Branch
	return &params, nil
}

func registerWebhook(app *bitriseio.AppService, inputReader io.Reader) error {
	var err error
	for i := 1; i <= 2; i++ {
		if err := app.RegisterWebhook(); err != nil {
			if e, ok := err.(*bitriseio.ErrorResponse); ok {
				if !httputil.IsUserFixable(e.Response.StatusCode) {
					return err
				}

				log.Errorf("Error registering webhook: %s", err)
				log.Warnf("Fix the error and hit enter to retry!")
				if _, err := bufio.NewReader(inputReader).ReadString('\n'); err != nil {
					return fmt.Errorf("failed to read line from input, error: %s", err)
				}
				continue
			}
		}
		return nil
	}
	return err
}

// Register ...
func Register(token string, progress Progress, inputReader io.Reader) error {
	log.Infof("Register")

	params, err := toRegistrationParams(progress)
	if err != nil {
		return err
	}

	log.Debugf("Provided params:\n%s", pretty.Object(params))

	client, err := bitriseio.NewClient(token)
	if err != nil {
		return err
	}
	app, err := client.Apps.Register(params.Repository)
	if err != nil {
		return err
	}
	if !params.Repository.IsPublic && params.SSHKey.AuthSSHPrivateKey != "" {
		if err := app.RegisterSSHKey(params.SSHKey, params.Repository.RepoURL); err != nil {
			return err
		}
	} else {
		log.Printf("Skipping SSH key registration.")
	}

	resp, err := app.RegisterFinish(params.Project)
	if err != nil {
		return err
	}

	log.Debugf(pretty.Object(resp))

	if err := app.UploadBitriseYML(params.BitriseYML); err != nil {
		return err
	}

	if params.RegisterWebhook {
		if resp.IsWebhookAutoRegSupported {
			if err := registerWebhook(app, inputReader); err != nil {
				log.Errorf("Failed to register webhook, error: %s", err)
			}
		} else {
			log.Errorf("Webhook registration is not possible right now, see options at: https://app.bitrise.io/app/%s#/code", app.Slug)
		}
		log.Warnf("Skipping webhook registration.")
	}

	if params.KeystorePth != "" {
		if err := app.UploadKeystore(params.KeystorePth, params.Keystore); err != nil {
			return err
		}
	}

	if len(params.CodesignIOS.certificates.Content) != 0 || len(params.CodesignIOS.provisioningProfiles) != 0 {
		// iOS codesigning files upload
		codesignIOSClient, err := bitrise.NewClient(token)
		if err != nil {
			return err
		}
		codesignIOSClient.SetSelectedAppSlug(app.Slug)

		if _, _, err := codesigndocBitriseio.UploadCodesigningFiles(codesignIOSClient, params.CodesignIOS.certificates, params.CodesignIOS.provisioningProfiles); err != nil {
			return err
		}
	} else if isIOSCodesign(params.Project.ProjectType) {
		log.Warnf(`To upload iOS code signing files, paste this script into a terminal on macOS and follow the instructions:
bash -l -c "$(curl -sfL https://raw.githubusercontent.com/bitrise-io/codesigndoc/master/_scripts/install_wrap.sh)"`)
	}

	if err := app.TriggerBuild(params.WorkflowID, params.Branch); err != nil {
		return err
	}

	log.Donef("Project created: https://app.bitrise.io/app/%s", app.Slug)
	return nil
}
