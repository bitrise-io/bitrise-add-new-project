package phases

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
)

// AddWebhook phase interrogates the user whether to create a webhook or not.
func AddWebhook() (bool, error) {
	fmt.Println()
	log.Infof("WEBHOOK SETUP")
	log.Printf("For automatic webhook setup for push and PR git events you need administrator rights for your repository")

	const (
		optionYes = "Yes"
		optionNo  = "No"
	)

	prompt := promptui.Select{
		Label: "Would you like us to register a webhook for you?",
		Items: []string{optionYes, optionNo},
		Templates: &promptui.SelectTemplates{
			Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
			Selected: "Auto register webhook: {{ . | green }}",
		},
	}

	_, answer, err := prompt.Run()
	if err != nil {
		return false, fmt.Errorf("scan user input: %s", err)
	}

	var registerWebhook bool
	if answer == optionYes {
		registerWebhook = true
	}

	return registerWebhook, nil
}
