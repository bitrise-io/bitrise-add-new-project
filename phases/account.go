package phases

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
)

type organizationData struct {
	Name string
	Slug string
}

type organizationsRespone struct {
	Data []organizationData
}

type meData struct {
	Username string
}

type meResponse struct {
	Data meData
}

func fetchOrgs(apiToken string) (*organizationsRespone, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/organizations", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server response: %s", res.Status)
	}

	var orgs organizationsRespone
	if err := json.NewDecoder(res.Body).Decode(&orgs); err != nil {
		return nil, err
	}

	return &orgs, nil
}

func fetchUser(apiToken string) (*meResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+apiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server response: %s", res.Status)
	}

	var me meResponse
	if err := json.NewDecoder(res.Body).Decode(&me); err != nil {
		return nil, err
	}

	return &me, nil
}

// Account returns the slug of the selected account. If the user selects
// the personal account, the slug is empty.
func Account(apiToken string, personal bool) (string, error) {
	user, err := fetchUser(apiToken)
	if err != nil {
		return "", fmt.Errorf("fetch authenticated user: %s", err)
	}

	if personal {
		log.Infof("CHOOSE ACCOUNT")
		log.Donef(colorstring.Greenf("Selected account: ") + user.Data.Username)
		fmt.Println()
		return "", nil
	}

	orgs, err := fetchOrgs(apiToken)
	if err != nil {
		return "", fmt.Errorf("fetch orgs for authenticated user: %s", err)
	}

	orgNameToSlug := map[string]string{}
	items := []string{user.Data.Username}
	for _, data := range orgs.Data {
		orgNameToSlug[data.Name] = data.Slug
		items = append(items, data.Name)
	}

	log.Infof("CHOOSE ACCOUNT")

	prompt := promptui.Select{
		Label: "Select account to use",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Selected: "Selected account: {{ . | green }}",
		},
	}

	_, acc, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("scan user input: %s", err)
	}

	fmt.Println()

	return orgNameToSlug[acc], nil
}
