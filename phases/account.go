package phases

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"net/http"

)

type organizationData struct{
	Name string
	Slug string
}

type organizationsResponse struct{
	Data []organizationData
}

func isValid(choice int, limit int ) bool {
	return choice >= 1 && choice <= limit
}

// Account ...
func Account(apiToken string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.bitrise.io/v0.1/organizations", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token " + apiToken)
	res, err := http.DefaultClient.Do(req)
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
	res, err = http.DefaultClient.Do(req)
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

	var choice int
	for !isValid(choice, len(options)) {
		fmt.Print("CHOOSE ACCOUNT: ")
	
		// scan for input
		reader := bufio.NewReader(os.Stdin)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading choice from stdin: %s", err)
			fmt.Println()
			continue
		}
		
		choice, err = strconv.Atoi(strings.TrimSpace(input))
		if err != nil {
			fmt.Printf("error reading choice from stdin: %s", err)
			fmt.Println()
			continue
		} else if !isValid(choice, len(options)) {
			fmt.Printf("invalid choice: %s", err)
			fmt.Println()
			continue
		} else {
			break
		}
	}
	

	fmt.Printf("your choice was %d", choice)
	fmt.Println()
	
	return options[choice].Slug, nil
}
