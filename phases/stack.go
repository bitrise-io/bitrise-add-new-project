package phases

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
)

// Based on: https://github.com/bitrise-io/bitrise-website/blob/master/config/available_stacks.yml
const (
	defaultMacOSStack = "osx-xcode-13.1.x"
	defaultLinuxStack = "linux-docker-android-20.04"
)

var defaultStacks = map[string]string{
	"ios":          defaultMacOSStack,
	"macos":        defaultMacOSStack,
	"android":      defaultLinuxStack,
	"cordova":      defaultMacOSStack,
	"ionic":        defaultMacOSStack,
	"react-native": defaultMacOSStack,
	"flutter":      defaultMacOSStack,
	"other":        defaultLinuxStack,
}

type availableStacksResponse map[string]interface{}

func fetchAvailableStacks(orgSlug string, apiToken string) ([]string, error) {
	var url string
	if orgSlug != "" {
		url = fmt.Sprintf("https://api.bitrise.io/v0.1/organizations/%s/available-stacks", orgSlug)
	} else {
		url = "https://api.bitrise.io/v0.1/me/available-stacks"
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
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

	var jsonMap availableStacksResponse
	if err := json.NewDecoder(res.Body).Decode(&jsonMap); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(jsonMap))
	for key := range jsonMap {
		keys = append(keys, key)
	}

	return keys, nil
}

// Stack returns the selected stack for the project or an error
// if something went wrong during stack autodetection.
func Stack(orgSlug string, apiToken string, projectType string) (string, error) {
	fmt.Println()
	log.Infof("SELECT STACK")
	stack := defaultStacks[projectType]

	availableStacks, err := fetchAvailableStacks(orgSlug, apiToken)
	if err != nil {
		return "", fmt.Errorf("Failed to fetch available stacks: %s", err)
	}

	if stack == "" {
		log.Warnf("Could not identify default stack for project. Falling back to manual stack selection.")

		prompt := promptui.Select{
			Label: "Please choose from the available stacks",
			Items: availableStacks,
			Templates: &promptui.SelectTemplates{
				Selected: "Stack: {{ . | green }}",
			},
		}

		_, stack, err = prompt.Run()
		if err != nil {
			return "", fmt.Errorf("scan user input: %s", err)
		}

		return stack, nil
	}

	systemReportURL := fmt.Sprintf("https://github.com/bitrise-io/bitrise.io/blob/master/system_reports/%s.log", stack)
	log.Printf("Project type: %s", colorstring.Green(projectType))
	log.Printf("Default stack for your project type: %s", colorstring.Green(stack))
	log.Printf("You can check the preinstalled tools at: %s", systemReportURL)

	const (
		optionYes = "Yes"
		optionNo  = "No, I will select the stack manually"
	)

	prompt := promptui.Select{
		Label: "Do you wish to keep this stack?",
		Items: []string{optionYes, optionNo},
		Templates: &promptui.SelectTemplates{
			Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
			Selected: "Keep default stack: {{ . | green }}",
		},
	}

	for {
		_, keep, err := prompt.Run()
		if err != nil {
			return "", fmt.Errorf("scan user input: %s", err)
		}

		if keep == optionYes {
			return stack, nil
		}

		stackPrompt := promptui.Select{
			Label: "Choose stack",
			Items: availableStacks,
			Templates: &promptui.SelectTemplates{
				Selected: "Stack: {{ . | green }}",
			},
		}
		_, stack, err = stackPrompt.Run()
		if err != nil {
			return "", fmt.Errorf("user input: %s", err)
		}

		return stack, nil
	}

	return "", fmt.Errorf("invalid state")
}
