package phases

import "github.com/bitrise-io/go-utils/log"

// AddWebhook phase interrogates the user whether to create a webhook or not.
func AddWebhook() (bool, error) {
	var registerWebhook bool

	log.Printf("WEBHOOK SETUP")
	log.Printf("To have Bitrise automatically start a build every time you push code into your repository you can set up a webhook at your code hosting service which will automatically trigger a build on Bitrise with the code you push to your repository.")
	log.Printf("We can automatically register a Webhook for you if you have administrator rights for this repository.")

	const (
		optionYes = "yes"
		optionNo  = "no"
	)

	(&option{
		title:        "Would you like us to register a webhook for you?",
		valueOptions: []string{optionYes, optionNo},
		action: func(answer string) *option {
			if answer == optionYes {
				registerWebhook = true
			}
			return nil
		},
	}).run()

	return registerWebhook, nil
}
