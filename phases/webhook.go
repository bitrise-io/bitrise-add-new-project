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
	log.Printf(`To let Bitrise automatically start a build every time you:
- push code
- open a pull request
into your repository, you can set up a webhook at your code hosting service.`)
	log.Printf("We can automatically register a Webhook for you if you have administrator rights for this repository.")

	const (
		optionYes = "yes"
		optionNo  = "no"
	)

	prompt := promptui.Select{
		Label: "Would you like us to register a webhook for you?",
		Items: []string{optionYes, optionNo},
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
