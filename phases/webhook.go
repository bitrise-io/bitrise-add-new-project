package phases

import "github.com/bitrise-io/go-utils/log"

// AddWebhook ...
func AddWebhook() (bool, error) {
	var registerWebhook bool

	log.Printf("WEBHOOK SETUP")
	log.Printf("To have Bitrise automatically start a build every time you push code into your repository you can set up a webhook at your code hosting service which will automatically trigger a build on Bitrise with the code you push to your repository.")
	log.Printf("We can automatically register a Webhook for you if you have administrator rights for this repository.")

	(&option{
		title:        "Would you like us to register a webhok fro you?",
		valueOptions: []string{"yes", "no"},
		action: func(answer string) *option {
			if answer == "yes" {
				registerWebhook = true
			}
			return nil
		},
	}).run()

	return registerWebhook, nil
}
