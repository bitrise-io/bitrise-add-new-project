package phases

import (
	"fmt"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
)

var defaultStacks = map[string]string{
	"xamarin":      "osx-vs4mac-stable",
	"cordova":      "osx-xcode-10.2.x",
	"react-native": "osx-xcode-10.2.x",
	"ionic":        "osx-xcode-10.2.x",
	"flutter":      "osx-xcode-10.2.x",
	"android":      "linux-docker-android",
	"macos":        "osx-xcode-10.2.x",
	"ios":          "osx-xcode-10.2.x",
}

var optionsStacks = []string{
	"linux-docker-android-lts",
	"linux-docker-android",
	"osx-vs4mac-beta",
	"osx-vs4mac-previous-stable",
	"osx-vs4mac-stable",
	"osx-xamarin-stable",
	"osx-xcode-10.0.x",
	"osx-xcode-10.1.x",
	"osx-xcode-10.2.x",
	"osx-xcode-8.3.x",
	"osx-xcode-9.2.x",
	"osx-xcode-9.4.x",
	"osx-xcode-edge",
}

// Stack returns the selected stack for the project or an error
// if something went wrong during stack autodetection.
func Stack(projectType string) (string, error) {
	fmt.Println()
	log.Infof("SELECT STACK")
	stack := defaultStacks[projectType]
	var err error

	if stack == "" {
		log.Warnf("Could not identify default stack for project. Falling back to manual stack selection.")

		prompt := promptui.Select{
			Label: "Please choose from the available stacks",
			Items: optionsStacks,
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
			Items: optionsStacks,
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
