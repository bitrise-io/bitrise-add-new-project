package phases

import (
	"fmt"
	"io/ioutil"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/log"
	"gopkg.in/yaml.v2"
)

var defaultStacks = map[string]string{
	"xamarin":      "osx-vs4mac-stable",
	"cordova":      "osx-vs4mac-stable",
	"react-native": "osx-vs4mac-stable",
	"ionic":        "osx-vs4mac-stable",
	"flutter":      "osx-vs4mac-stable",
	"android":      "linux-docker-android",
	"macos":        "osx-xcode-10.0",
	"ios":          "osx-xcode-10.0",
	"other":        "",
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

// Stack ...
func Stack(bitriseYMLPath string) (string, error) {

	data, err := ioutil.ReadFile("bitrise.yml")
	if err != nil {
		log.Errorf("read bitrise yml: %s", err)
		return "", nil
	}

	var m models.BitriseDataModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		log.Errorf("unmarshal bitrise yml: %s", err)
		return "", nil
	}

	projectType := m.ProjectType
	if projectType == "" {
		projectType = "other"
	}

	var stack string
	var manualStackSelection = option{
		title:        "Please choose from the available stacks",
		valueOptions: optionsStacks,
		action: func(answer string) *option {
			stack = answer
			return nil
		},
	}

	if stack = defaultStacks[projectType]; stack != "" {
		systemReportURL := fmt.Sprintf("https://github.com/bitrise-io/bitrise.io/blob/master/system_reports/%s.log", stack)
		log.Printf("An %d project has been detected based on the provided bitrise.yml (%s)", stack, bitriseYMLPath)
		log.Printf("The default stack for your project type is %s. You can check the preinstalled tools at %s", stack, systemReportURL)

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
		return stack, nil
	}

	log.Warnf("Could not identify default stack for project. Falling back to manual stack selection.")
	(&manualStackSelection).run()

	return stack, nil
}
