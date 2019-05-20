package phases

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
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

func checkBranch(inputReader io.Reader, currentBranch func() (string, error)) error {
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
			scanner := bufio.NewScanner(inputReader)
			if scanner.Scan() {
				_ = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read line from input, error: %s", err)
			}
			continue
		}
		break
	}
	return nil
}

func openFile(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to open file (%s), error: %s", filePath, err)
		}
		return file, nil
	}
	return nil, nil
}

func parseDSLFile(input io.Reader) (models.BitriseDataModel, error) {
	var decodedDSL models.BitriseDataModel
	if err := yaml.NewDecoder(input).Decode(&decodedDSL); err != nil {
		return models.BitriseDataModel{}, fmt.Errorf("failed to parse bitrise.yml, error: %s", err)
	}
	return decodedDSL, nil
}

// ReadExistingDSLFile reads and parse a given bitrise.yml from a path
func ReadExistingDSLFile(filePath string) (models.BitriseDataModel, bool, error) {
	DSLFile, err := openFile(filePath)
	if err != nil {
		return models.BitriseDataModel{}, false, err
	} else if DSLFile == nil {
		log.Infof("No bitrise.yml file found in the current working directory.")
		return models.BitriseDataModel{}, false, nil
	}

	DSL, err := parseDSLFile(DSLFile)
	return DSL, true, err
}

func selectDSLFile(inputReader io.Reader) (models.BitriseDataModel, bool, error) {
	for {
		runScanner, err := goinp.AskForBoolFromReaderWithDefaultValue("Input bitrise.yml path manually? (Will generate otherwise.)", false, inputReader)
		if err != nil {
			log.Errorf("%s", err)
			return models.BitriseDataModel{}, false, err
		} else if runScanner {
			return models.BitriseDataModel{}, true, nil
		}

		path, err := goinp.AskForPathFromReader("Enter the path of your bitrise.yml file (you can also drag & drop the file here)", inputReader)
		if err != nil {
			log.Errorf("%s", err)
			return models.BitriseDataModel{}, false, err
		}

		DSLFile, err := openFile(path)
		if err != nil {
			return models.BitriseDataModel{}, false, err
		} else if DSLFile == nil {
			log.Warnf("File (%s) does not exist.", path)
			continue
		}

		decodedDSL, err := parseDSLFile(DSLFile)
		if err != nil {
			log.Warnf("Failed to parse bitrise.yml, error: %s", err)
			continue
		}
		return decodedDSL, false, nil
	}
}

func runScanner() (models.BitriseDataModel, error) {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return models.BitriseDataModel{}, err
	}
	DSL, err := scanAndAskForProjectDetails(currentDir)
	if err != nil {
		return models.BitriseDataModel{}, err
	}
	return DSL, nil
}

func selectWorkflow(buildDSL models.BitriseDataModel, inputReader io.Reader) (string, error) {
	var workflows []string
	for workflow := range buildDSL.Workflows {
		workflows = append(workflows, workflow)
	}

	workflow, err := goinp.SelectFromStringsFromReaderWithDefault("Select workflow to run in the first build:", 0, workflows, inputReader)
	if err != nil {
		log.Errorf("%s", err)
		return "", err
	}
	return workflow, nil
}

// BitriseYML ...
func BitriseYML(searchDir string) (models.BitriseDataModel, string, error) {
	potentialDSLFilePath := filepath.Join(searchDir, bitriseYMLName)
	DSL, found, err := ReadExistingDSLFile(potentialDSLFilePath)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to read existing bitrise.yml, error: %s", err)
	} else if found {
		workflow, err := selectWorkflow(DSL, os.Stdin)
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to select workflow, error: %s", err)
		}
		return DSL, workflow, nil
	}

	DSL, cancelled, err := selectDSLFile(os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to select bitrise.yml, error: %s", err)
	} else if !cancelled {
		workflow, err := selectWorkflow(DSL, os.Stdin)
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to select workflow, error: %s", err)
		}
		return DSL, workflow, nil
	}

	err = checkBranch(os.Stdin, currentBranch)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to check repository branch: %s", err)
	}

	DSL, err = runScanner()
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to run repository scan, error: %s", err)
	}
	workflow, err := selectWorkflow(DSL, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to select workflow, error: %s", err)
	}
	return DSL, workflow, nil
}
