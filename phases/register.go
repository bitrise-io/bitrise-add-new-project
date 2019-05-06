package phases

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bitrise-io/go-utils/log"
)

const baseURL = "https://app.bitrise.io"

func startAppRegistration() (string, error) {
	return "", nil
}

func registerSSHKey() error {
	return nil
}

func registerWebhook(appSlug string, apiToken string) error {
	url := fmt.Sprintf("%s/app/%s/register-webhook.json", baseURL, appSlug)
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		// todo
	}

	request.Header.Add("Authorization", "token "+ apiToken)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		// todo
	}

	if resp.StatusCode == 400 {
		log.Errorf("Error registering webhook")
		log.Warnf("Make sure you have the required access rights to the repository and that you enabled git provider integration jfor your Bitrise account!")
		
		reader := bufio.NewReader(os.Stdin)

		for {
			_, _ = reader.ReadString('\n')
			if err := registerWebhook(appSlug, apiToken); err == nil {
				return nil
			}
		}
	}

	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// todo
		}
		var m map[string]string

		if err := json.Unmarshal(data, &m); err != nil {
			// todo
		}

		log.Errorf("Error registering webhook: %s %s", resp.Status, m["error_msg"])
	}

	log.Successf("Webhook registered")
	return nil
}

func finishAppRegistration() error {
	return nil
}

// Register ...
func Register(progress Progress, apiToken string) error {
	fmt.Println("Register")

	appSlug, err := startAppRegistration()
	if err != nil {
		return err
	}

	if err := registerSSHKey(); err != nil {
		return err
	}

	if *progress.AddWebhook {
		if err := registerWebhook(appSlug, apiToken); err != nil {
			return err
		}
	}

	if err := finishAppRegistration(); err != nil {
		return err
	}
	return nil
}
