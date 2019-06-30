package phases

import (
	"fmt"

	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
)

const (
	optPrivate = "Private"
	optPublic  = "Public"
)

// IsPublic returns the whether the Bitrise project
// should be public or not.
func IsPublic() (bool, error) {
	items := []string{optPrivate, optPublic}

	fmt.Println()
	log.Infof("SET THE PRIVACY OF THE APP")
	prompt := promptui.Select{
		Label: "Choose who can see you app logs and configs on bitrise.io",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Selected: "Selected visibility: {{ . | green }}",
		},
	}

	_, visibility, err := prompt.Run()
	if err != nil {
		return false, fmt.Errorf("scan user input: %s", err)
	}

	return visibility == optPublic, nil
}
