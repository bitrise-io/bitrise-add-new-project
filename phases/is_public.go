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
	log.Infof("SET PRIVACY OF THE PROJECT")
	prompt := promptui.Select{
		Label: "Select privacy",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Selected: "Selected privacy: {{ . | green }}",
		},
	}

	_, visibility, err := prompt.Run()
	if err != nil {
		return false, fmt.Errorf("scan user input: %s", err)
	}

	return visibility == optPublic, nil
}
