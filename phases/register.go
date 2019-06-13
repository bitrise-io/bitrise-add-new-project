package phases

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/bitrise-io/bitrise-add-new-project/bitriseio"
	"github.com/bitrise-io/bitrise-add-new-project/httputil"
	codesigndocBitriseio "github.com/bitrise-io/codesigndoc/bitriseio"
	"github.com/bitrise-io/codesigndoc/bitriseio/bitrise"
	"github.com/bitrise-io/go-utils/fileutil"
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

	privateKey, err := fileutil.ReadStringFromFile(progress.SSHPrivateKeyPth)
	if err != nil {
		return nil, fmt.Errorf("SSH private key read failed: %s", err)
	}
	// privateKey = strings.Replace(privateKey, "\n", "\\n", -1)
	privateKey = strings.TrimSuffix(privateKey, "\n")
	privateKey = strings.Replace(privateKey, "OPENSSH", "RSA", -1)
	progress.RegisterSSHKey = false

	var publicKey string
	if progress.RegisterSSHKey {
		var err error
		publicKey, err = fileutil.ReadStringFromFile(progress.SSHPublicKeyPth)
		if err != nil {
			return nil, fmt.Errorf("SSH public key read failed: %s", err)
		}
	}

	params := CreateProjectParams{}
	params.Repository = bitriseio.RegisterParams{
		GitOwner:    progress.RepoURL.Owner,
		GitRepoSlug: progress.RepoURL.Slug,
		IsPublic:    progress.Public,
		Provider:    progress.RepoURL.Provider,
		RepoURL:     progress.RepoURL.URL,
	}
	params.RegisterWebhook = progress.AddWebhook
	params.SSHKey = bitriseio.RegisterSSHKeyParams{
		AuthSSHPrivateKey:                privateKey,
		AuthSSHPublicKey:                 publicKey,
		IsRegisterKeyIntoProviderService: progress.RegisterSSHKey,
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
	return &params, nil
}

// Register ...
func Register(token string, progress Progress, inputReader io.Reader) error {
	log.Infof("Register")

	params, err := toRegistrationParams(progress)
	if err != nil {
		return err
	}

	log.Printf("Provided params:\n%s", pretty.Object(params))

	client, err := bitriseio.NewClient(token)
	if err != nil {
		return err
	}
	app, err := client.Apps.Register(params.Repository)
	if err != nil {
		return err
	}
	if !params.Repository.IsPublic {
		if err := app.RegisterSSHKey(params.SSHKey); err != nil {
			return err
		}
	}

	if params.RegisterWebhook {
		if err := registerWebhook(app, inputReader); err != nil {
			log.Errorf("Failed to register webhook, error: %s", err)
			log.Warnf("Skipping webhook registration.")
		}
	}

	resp, err := app.RegisterFinish(params.Project)
	if err != nil {
		return err
	}

	fmt.Println(pretty.Object(resp))

	if err := app.UploadBitriseYML(params.BitriseYML); err != nil {
		return err
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
	} else {
		log.Warnf(`To upload iOS code signing files, paste this script into your terminal and follow the instructions:
bash -l -c "$(curl -sfL https://raw.githubusercontent.com/bitrise-io/codesigndoc/master/_scripts/install_wrap.sh)"`)
	}

	if err := app.TriggerBuild(params.WorkflowID); err != nil {
		return err
	}

	log.Donef("Project created: https://app.bitrise.io/app/%s", app.Slug)
	return nil
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
