package phases

import (
	"github.com/bitrise-io/bitrise-add-new-project/bitrise"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/xcode-project/pretty"
)

func toRegistrationParams(progress Progress) (params bitrise.CreateProjectParams, err error) {
	params.Repository = bitrise.RegisterParams{
		GitOwner:    *progress.RepoOwner,
		GitRepoSlug: *progress.RepoSlug,
		IsPublic:    *progress.Public,
		Provider:    *progress.RepoProvider,
		RepoURL:     *progress.RepoURL,
		Type:        *progress.RepoType,
	}
	params.RegisterWebhook = *progress.AddWebhook
	params.SSHKey = bitrise.RegisterSSHKeyParams{
		AuthSSHPrivateKey:                *progress.PrivateKey,
		AuthSSHPublicKey:                 "",
		IsRegisterKeyIntoProviderService: true,
	}
	params.Project = bitrise.RegisterFinishParams{
		Config:           *progress.BitriseYML,
		Envs:             nil,
		Mode:             "manual",
		OrganizationSlug: *progress.Account,
		ProjectType:      "",
		StackID:          *progress.Stack,
	}
	return
}

// Register ...
func Register(token string, progress Progress) error {
	log.Infof("Register")

	params, err := toRegistrationParams(progress)
	if err != nil {
		return err
	}

	log.Printf("Provided params: %s", pretty.Object(params))

	slug, err := bitrise.CreateProject(token, params)
	if err != nil {
		return err
	}

	log.Donef("Project created: https://app.bitrise.io/app/%s", slug)
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
