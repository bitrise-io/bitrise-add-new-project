package phases

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

type organizationData map[string]interface{}

func isValid(choice int, limit int) bool {
	return choice >= 1 && choice <= limit
}

// Account returns the slug of the selected account. If the user selects
// the personal account, the slug is empty.
func Account(apiToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/organizations", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("fetch orgs: %s", res.Status)
	}

	var m struct {
		Data []organizationData
	}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", err
	}

	req, err = http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/me", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+apiToken)
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("fetch user: %s", res.Status)
	}

	var u struct {
		Data struct {
			Username string
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return "", err
	}

	options := append([]organizationData{organizationData{"name": u.Data.Username}}, m.Data...)

	log.Infof("ACCOUNT OPTIONS")
	log.Infof("===============")
	for i, opt := range options {
		log.Printf("%d) %s", i+1, opt["name"])
	}

	for {
		log.Warnf("CHOOSE ACCOUNT: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf("error reading choice from stdin: %s", err)
			continue
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			log.Errorf("error reading choice from stdin: %s", err)
			continue
		}
		if !isValid(choice, len(options)) {
			log.Errorf("invalid choice")
			continue
		}
		if choice == 1 {
			// own user selected
			return "", nil
		}

		// organization selected
		return options[choice-1]["slug"].(string), nil
	}
}
