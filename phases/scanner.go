package phases

import (
	"fmt"

	"github.com/bitrise-io/bitrise-init/scanner"
	"github.com/bitrise-io/bitrise-init/scanners"
	"github.com/bitrise-io/bitrise/models"
	"github.com/go-yaml/yaml"
)

func scanAndAskForProjectDetails(searchDir string) (models.BitriseDataModel, error) {
	result, found := scanner.GenerateScanResult(searchDir)
	if !found {
		return models.BitriseDataModel{}, fmt.Errorf("no known project type found")
	}

	workflow, err := scanner.AskForConfig(result)
	if err != nil {
		return models.BitriseDataModel{}, err
	}
	return workflow, nil
}

func manualConfig() (models.BitriseDataModel, error) {
	scanResult, err := scanner.ManualConfig()
	if err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to create empty config, error: %s", err)
	}

	customConfigs, ok := scanResult.ScannerToBitriseConfigMap[scanners.CustomProjectType]
	if !ok {
		return models.BitriseDataModel{}, fmt.Errorf("no CustomProjectType found found, error: %s", err)
	}

	customConfigStr, ok := customConfigs[scanners.CustomConfigName]
	if !ok {
		return models.BitriseDataModel{}, fmt.Errorf("no CustomConfig found, error: %s", err)
	}

	var customConfig models.BitriseDataModel
	if err := yaml.Unmarshal([]byte(customConfigStr), &customConfig); err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("no default empty config found, error: %s", err)
	}

	return customConfig, nil
}
