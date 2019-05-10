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
		return fmt.Errorf("create POST %s request: %s", url, err)
	}

	request.Header.Add("Authorization", "token "+apiToken)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("send POST %s request: %s", url, err)
	}

	if resp.StatusCode == 400 {
		log.Errorf("Error registering webhook")
		log.Warnf("Make sure you have the required access rights to the repository and that you enabled git provider integration jfor your Bitrise account!")
		log.Warnf("Please fix your configuration and hit enter to try again!")

		reader := bufio.NewReader(os.Stdin)

		for {
			if _, err = reader.ReadString('\n'); err != nil {
				log.Errorf("Error reading user input")
				continue
			}
			if err := registerWebhook(appSlug, apiToken); err != nil {
				log.Errorf("Error registering webhook: %s", err)
				log.Warnf("Please fix your configuration and hit enter to try again!")
				continue
			}
			return nil
		}
	}

	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read create webhook response: %s", err)
		}
		var m map[string]string

		if err := json.Unmarshal(data, &m); err != nil {
			return fmt.Errorf("unmarshal create webhook response: %s", err)
		}

		log.Errorf("Error registering webhook: %s %s", resp.Status, m["error_msg"])
		return fmt.Errorf("server error registering webhook: %s %s", resp.Status, m["error_msg"])
	}

	log.Successf("Webhook registered")
	return nil
}

func finishAppRegistration() error {
	return nil
}

// Register sends the data to the Bitrise servers, effectively creating the
// application.
func Register(progress Progress, apiToken string) error {

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
