package phases

import (
	"io/ioutil"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/log"
	"gopkg.in/yaml.v2"
)

var defaultStacks map[string]string = map[string]string{
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

	var stack string

	(&option{
		title:        "Choose stack selection mode",
		valueOptions: []string{"auto", "manual"},
		action: func(answer string) *option {
			if answer == "auto" {

				data, err := ioutil.ReadFile("bitrise.yml")
				if err != nil {
					log.Errorf("read bitrise yml: %s", err)
					return nil
				}

				var m models.BitriseDataModel
				if err := yaml.Unmarshal(data, &m); err != nil {
					log.Errorf("unmarshal bitrise yml: %s", err)
					return nil
				}

				projectType := m.ProjectType
				if projectType == "" {
					projectType = "other"
				}

				if stack = defaultStacks[projectType]; stack != "" {
					return nil
				}

				log.Warnf("Could not identify default stack for project type (%s). Falling back to manual stack selection.", projectType)
				answer = "manual"
			}

			if answer == "manual" {
				(&option{
					title:        "Available stacks",
					valueOptions: optionsStacks,
					action: func(answer string) *option {
						stack = answer
						return nil
					},
				}).run()
			}

			return nil
		}}).run()

	log.Successf(stack)
	return stack, nil
}
