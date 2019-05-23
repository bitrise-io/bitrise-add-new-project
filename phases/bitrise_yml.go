package phases

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanner"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

const bitriseYMLName = "bitrise.yml"

func currentBranch() (string, error) {
	cmd := command.New("git", "symbolic-ref", "HEAD")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", errors.Wrap(fmt.Errorf("failed to run command: %s, error: %s", cmd.PrintableCommandArgs(), err), out)
	}
	return strings.TrimPrefix(out, "refs/heads/"), nil
}

func checkBranch(inputReader io.Reader) error {
	var branch string
	for {
		var err error
		if branch, err = currentBranch(); err != nil {
			return fmt.Errorf("failed to get current branch, error: %s", err)
		}
		msg := fmt.Sprintf("The current branch is: %s. Do you want to run the scanner for this branch?", branch)
		useCurrentBranch, err := goinp.AskForBoolFromReaderWithDefaultValue(msg, true, inputReader)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		if !useCurrentBranch {
			log.Printf("Checkout a different branch then press Enter.")
			if _, err := bufio.NewReader(inputReader).ReadString('\n'); err != nil {
				return fmt.Errorf("failed to read line from input, error: %s", err)
			}
			continue
		}
		break
	}
	return nil
}

// ParseBitriseYMLFile parses a bitrise.yml and returns a data model
func ParseBitriseYMLFile(input io.Reader) (models.BitriseDataModel, error) {
	var decodedBitriseYML models.BitriseDataModel
	if err := yaml.NewDecoder(input).Decode(&decodedBitriseYML); err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to parse bitrise.yml, error: %s", err)
	}
	return decodedBitriseYML, nil
}

func selectBitriseYMLFile(inputReader io.Reader) (models.BitriseDataModel, bool, error) {
	for {
		inputPathManually, err := goinp.AskForBoolFromReaderWithDefaultValue("Input bitrise.yml path manually? (Will generate otherwise.)", false, inputReader)
		if err != nil {
			log.Errorf("%s", err)
			return models.BitriseDataModel{}, false, err
		} else if !inputPathManually {
			return models.BitriseDataModel{}, true, nil
		}

		filePath, err := goinp.AskForPathFromReader("Enter the path of your bitrise.yml file (you can also drag & drop the file here)", inputReader)
		if err != nil {
			log.Errorf("%s", err)
			return models.BitriseDataModel{}, false, err
		}

		BitriseYMLFile, err := os.Open(filePath)
		defer func() {
			if err := BitriseYMLFile.Close(); err != nil {
				log.Warnf("failed to close file, error: %s", err)
			}
		}()
		if err != nil {
			if !os.IsNotExist(err) {
				return models.BitriseDataModel{}, false, fmt.Errorf("failed to open file (%s), error: %s", filePath, err)
			}
			log.Warnf("File (%s) does not exist.", filePath)
			continue
		}

		decodedBitriseYML, err := ParseBitriseYMLFile(BitriseYMLFile)
		if err != nil {
			log.Warnf("Failed to parse bitrise.yml, error: %s", err)
			continue
		}
		return decodedBitriseYML, false, nil
	}
}

func selectWorkflow(buildBitriseYML models.BitriseDataModel, inputReader io.Reader) (string, error) {
	if len(buildBitriseYML.Workflows) == 0 {
		return "", fmt.Errorf("no workflows found in bitrise.yml")
	}

	const defaultWorkflowName = "primary"
	if _, contains := buildBitriseYML.Workflows[defaultWorkflowName]; contains {
		return defaultWorkflowName, nil
	}

	var workflows []string
	for workflow := range buildBitriseYML.Workflows {
		workflows = append(workflows, workflow)
	}

	if len(workflows) == 1 {
		log.Infof("Selecting workflow: %s", workflows[0])
		return workflows[0], nil
	}

	workflow, err := goinp.SelectFromStringsFromReaderWithDefault("Select workflow to run in the first build:", 1, workflows, inputReader)
	if err != nil {
		log.Errorf("%s", err)
		return "", err
	}
	return workflow, nil
}

func getBitriseYML(searchDir string, inputReader io.Reader) (models.BitriseDataModel, error) {
	potentialBitriseYMLFilePath := filepath.Join(searchDir, bitriseYMLName)

	if exist, err := pathutil.IsPathExists(potentialBitriseYMLFilePath); err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to check if file (%s) exists, error: %s", potentialBitriseYMLFilePath, err)
	} else if exist {
		log.Infof("Found bitrise.yml in current directory.")
		file, err := os.Open(potentialBitriseYMLFilePath)
		defer func() {
			if err := file.Close(); err != nil {
				log.Warnf("failed to close file, error: %s", err)
			}
		}()
		if err != nil && !os.IsNotExist(err) {
			return models.BitriseDataModel{}, fmt.Errorf("failed to open file (%s), error: %s", potentialBitriseYMLFilePath, err)
		}
		bitriseYML, err := ParseBitriseYMLFile(file)
		if err != nil {
			return models.BitriseDataModel{}, fmt.Errorf("failed to parse bitrise.yml, error: %s", err)
		}
		return bitriseYML, nil
	}

	bitriseYML, cancelled, err := selectBitriseYMLFile(inputReader)
	if err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to select bitrise.yml, error: %s", err)
	} else if !cancelled {
		return bitriseYML, nil
	}

	err = checkBranch(inputReader)
	if err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to check repository branch: %s", err)
	}

	scanResult, found := scanner.GenerateScanResult(searchDir)
	if !found {
		log.Infof("Projects not found in repository. BitriseYMLure project manually.")
		scanResult, err = scanner.ManualConfig()
		if err != nil {
			return models.BitriseDataModel{}, fmt.Errorf("failed to get manual bitriseYMLurations, error: %s", err)
		}
	} else {
		log.Infof("Projects found in repository.")
	}
	bitriseYML, err = scanner.AskForConfig(scanResult)
	if err != nil {
		log.Errorf("%s", err)
		return models.BitriseDataModel{}, fmt.Errorf("failed to get exact bitriseYMLuration from scanner result, error: %s", err)
	}
	return bitriseYML, nil
}

// BitriseYML ...
func BitriseYML(searchDir string) (models.BitriseDataModel, string, error) {
	bitriseYML, err := getBitriseYML(searchDir, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", err
	}

	workflow, err := selectWorkflow(bitriseYML, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to select workflow, error: %s", err)
	}
	return bitriseYML, workflow, nil
}
