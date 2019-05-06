package phases

import (
	"io/ioutil"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/log"
	"gopkg.in/yaml.v2"
)

// Stack ...
func Stack(bitriseYMLPath string) (string, error) {

	(&option{
		title:        "Choose stack selection mode",
		valueOptions: []string{"auto", "manual"},
		action: func(answer string) *option {
			switch answer {
			case "auto":

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

				if m.ProjectType == "" {
					log.Warnf("Could not identify default stack: %s contains no project_type property. Falling back to manual stack selection.", bitriseYMLPath)
					projectType := "other"
					log.Printf(projectType)
					return nil
				}

				projectType := m.ProjectType

				log.Printf(projectType)

			case "manual":
				log.Printf("manual run selected")
			}

			return nil
		}}).run()

	return "", nil
}
