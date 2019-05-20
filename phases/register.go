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

var baseURL = "https://app.bitrise.io"

var (
	webhookAttemptCount int
	webhookAttemptMax   = 3
)

func startAppRegistration() (string, error) {
	return "", nil
}

func registerSSHKey() error {
	return nil
}

func registerWebhook(appSlug string, apiToken string) error {
	webhookAttemptCount = 0
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Failed to close response body, error: %s", err)
		}
	}()

	if resp.StatusCode == http.StatusBadRequest {
		log.Errorf("Error registering webhook")
		log.Warnf("Make sure you have the required access rights to the repository and that you enabled git provider integration for your Bitrise account!")
		log.Warnf("Please fix your configuration and hit enter to try again!")

		reader := bufio.NewReader(os.Stdin)

		for {
			if webhookAttemptCount == webhookAttemptMax {
				return fmt.Errorf("maximum number of retries reached")
			}
			if _, err = reader.ReadString('\n'); err != nil {
				log.Errorf("Error reading user input")
				webhookAttemptCount++
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
