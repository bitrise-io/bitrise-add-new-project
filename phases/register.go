package phases

import (
	"fmt"

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
}

func toRegistrationParams(progress Progress) (*CreateProjectParams, error) {
	bitriseYML, err := yaml.Marshal(*progress.BitriseYML)
	if err != nil {
		return nil, err
	}
	str := string(bitriseYML)

	params := CreateProjectParams{}
	params.Repository = bitriseio.RegisterParams{
		GitOwner:    *progress.RepoOwner,
		GitRepoSlug: *progress.RepoSlug,
		IsPublic:    *progress.Public,
		Provider:    *progress.RepoProvider,
		RepoURL:     *progress.RepoURL,
		Type:        *progress.RepoType,
	}
	params.RegisterWebhook = *progress.AddWebhook
	params.SSHKey = bitriseio.RegisterSSHKeyParams{
		AuthSSHPrivateKey:                *progress.PrivateKey,
		AuthSSHPublicKey:                 "",
		IsRegisterKeyIntoProviderService: true,
	}
	params.Project = bitriseio.RegisterFinishParams{
		Config:           str,
		Envs:             nil,
		Mode:             "manual",
		OrganizationSlug: *progress.Account,
		ProjectType:      "",
		StackID:          *progress.Stack,
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

	if err := service.TriggerBuild(slug, params.TriggerBuild); err != nil {
		return err
	}

	log.Donef("Project created: https://app.bitriseio.io/app/%s", slug)
	return nil
}

/*
func performRegisterWebhookRequest(appSlug string, apiToken string) (*http.Response, error) {
	url := fmt.Sprintf("%s/app/%s/register-webhook.json", baseURL, appSlug)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create POST %s request: %s", url, err)
	}

	request.Header.Add("Authorization", "token "+apiToken)

	var resp *http.Response
	if err := retry.Times(3).Wait(5 * time.Second).Try(func(attempt uint) error {
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			log.Warnf("Could not POST to %s: %s -- will retry", url, err)
			return err
		}
		return nil
	}); err != nil {
		log.Warnf("Retry limit reached for sending create webhook request")
		return nil, fmt.Errorf("send POST %s request: %s", url, err)
	}
	return resp, nil
}

func registerWebhook(appSlug string, apiToken string) error {
	resp, err := performRegisterWebhookRequest(appSlug, apiToken)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close response body, error: %s", err)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for resp.StatusCode == http.StatusBadRequest {
		log.Errorf("Error registering webhook")
		log.Warnf("Make sure you have the required access rights to the repository and that you enabled git provider integration for your Bitrise account!")
		log.Warnf("Please fix your configuration and hit enter to try again!")

		if _, err = reader.ReadString('\n'); err != nil {
			log.Errorf("Error reading user input")
			continue
		}

		resp, err = performRegisterWebhookRequest(appSlug, apiToken)
		if err != nil {
			return err
		}
	}

	if resp.StatusCode == http.StatusOK {
		log.Successf("Webhook registered")
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read create webhook response: %s", err)
	}
	var m map[string]string

	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("unmarshal create webhook response (%s): %s", string(data), err)
	}

	return fmt.Errorf("server error registering webhook: %s %s", resp.Status, m["error_msg"])
}
*/
