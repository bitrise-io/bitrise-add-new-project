package phases

import (
	"fmt"

	"github.com/bitrise-io/bitrise-init/stack"
	"github.com/bitrise-io/go-utils/log"
)

// Stack returns the selected stack for the project or an error
// if something went wrong during stack auto-detection.
func Stack(projectType string) (string, error) {
	selectedStack := stack.DefaultStacks[projectType]
	var manualStackSelection = option{
		title:        "Please choose from the available stacks",
		valueOptions: stack.OptionsStacks,
		action: func(answer string) *option {
			selectedStack = answer
			return nil
		},
	}

	if selectedStack == "" {
		log.Warnf("Could not identify default stack for project. Falling back to manual stack selection.")
		(&manualStackSelection).run()

		return selectedStack, nil
	}

	systemReportURL := fmt.Sprintf("https://github.com/bitrise-io/bitrise.io/blob/master/system_reports/%s.log", selectedStack)
	log.Printf("A(n) %s project has been detected based on the bitrise.yml", projectType)
	log.Printf("The default stack for your project type is %s. You can check the preinstalled tools at %s", selectedStack, systemReportURL)

	const (
		optionYes = "Yes"
		optionNo  = "No, I will select the stack manually"
	)
	(&option{
		title:        "Do you wish to keep this stack?",
		valueOptions: []string{optionYes, optionNo},
		action: func(answer string) *option {
			if answer == optionNo {
				log.Printf("Bitrise stack infos: https://github.com/bitrise-io/bitrise.io/tree/master/system_reports")
				(&manualStackSelection).run()
			}

			return nil
		},
	}).run()
	return selectedStack, nil

}
