package phases

import (
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"time"
	"net/http"

)

type organizationData struct{
	Name string
	Slug string
}

type organizationsResponse struct{
	Data []organizationData
}

// Account ...
func Account(apiToken string) (string, error) {
	fmt.Println("SetAccount")
	req, err := http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/organizations", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token " + apiToken)
	c := http.Client{
		Timeout: 3 * time.Second,
	}
	res, err := c.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("fetch orgs: %s", res.Status)
	}

	m := organizationsResponse{}
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", err
	}

	// display orgs
	fmt.Println(m)
	// scan for input

	return "", nil
}
