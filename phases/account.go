package phases

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	
	req, err = http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/me", nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "token " + apiToken)
	res, err = c.Do(req)
	if err != nil {
		return "", err
	}
	
	if res.StatusCode != 200 {
		return "", fmt.Errorf("fetch user: %s", res.Status)
	}
	
	u := struct{
		Data struct{
			Username string
		}
	}{}
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		return "", err
	}

	fmt.Println(u)
	
	options := []organizationData{organizationData{Name: u.Data.Username}}
	options = append(options, m.Data...)
	
	// display orgs
	fmt.Println("ACCOUNT OPTIONS")
	for i, opt := range options {
		fmt.Printf("%d) %s", i + 1, opt.Name)
		fmt.Println()
	}
	fmt.Print("CHOOSE ACCOUNT: ")
	
	// scan for input
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		// todo
	}
	
	choice, err := strconv.Atoi(input)
	if err != nil {
		// todo
	}

	fmt.Printf("your choice was %s", options[choice].Name)
	fmt.Println()
	
	return options[choice].Slug, nil
}
