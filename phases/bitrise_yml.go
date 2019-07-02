package phases

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-init/scanner"
	"github.com/bitrise-io/bitrise/bitrise"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/manifoldco/promptui"
	"gopkg.in/src-d/go-git.v4"
)

const bitriseYMLName = "bitrise.yml"

type branchConfiguration struct {
	local    string
	tracking string
	remote   string
}

func askBranch(currentBranch string, inputReader io.Reader) (string, error) {
	const msg = "Which branch would you like to be the default?"
	return goinp.AskForStringFromReaderWithDefault(msg, currentBranch, inputReader)
}

func currentBranch(searchDir string) (branchConfiguration, error) {
	// Open local git repository
	repo, err := git.PlainOpen(searchDir)
	if err != nil {
		return branchConfiguration{}, fmt.Errorf("failed to open git repository (%s), error: %s", searchDir, err)
	}
	log.Debugf("Opened git repository: %s", searchDir)

	head, err := repo.Head()
	if err != nil {
		return branchConfiguration{}, err
	}

	config, err := repo.Config()
	if err != nil {
		return branchConfiguration{}, err
	}

	branch := branchConfiguration{
		local: head.Name().Short(),
	}

	if config.Branches[branch.local] != nil {
		if err := config.Branches[branch.local].Validate(); err != nil {
			return branchConfiguration{}, fmt.Errorf("invalid tracking branch configuration, error: %s", err)
		}
		branch.tracking = config.Branches[branch.local].Name
		branch.remote = config.Branches[branch.local].Remote
	}

	return branch, nil
}

func checkBranch(searchDir string, inputReader io.Reader) (string, error) {
	var branch branchConfiguration
	for {
		var err error
		if branch, err = currentBranch(searchDir); err != nil {
			return "", fmt.Errorf("failed to get current branch, error: %s", err)
		}

		log.Printf("The current branch is: %s (tracking: %s %s).", colorstring.Green(branch.local), branch.remote, branch.tracking)
		if branch.tracking == "" {
			log.Errorf("No tracking branch is set for the current branch. Check out an other branch then press Enter.")
			if _, err := bufio.NewReader(inputReader).ReadString('\n'); err != nil {
				return "", fmt.Errorf("failed to read line from input, error: %s", err)
			}
			continue
		}

		prompt := promptui.Select{
			Label: "Do you want to run the scanner for this branch?",
			Items: []string{"Yes", "No"},
			Templates: &promptui.SelectTemplates{
				Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
				Selected: "Run the scanner on the current branch: {{ . | green }}",
			},
		}

		_, answer, err := prompt.Run()
		if err != nil {
			return "", err
		}

		if answer == "no" {
			log.Printf("Check out an other branch then press Enter.")
			if _, err := bufio.NewReader(inputReader).ReadString('\n'); err != nil {
				return "", fmt.Errorf("failed to read line from input, error: %s", err)
			}
			continue
		}
		break
	}
	return branch.tracking, nil
}

// ParseBitriseYMLFile parses a bitrise.yml and returns a data model
func ParseBitriseYMLFile(inputReader io.Reader) (models.BitriseDataModel, []string, error) {
	content, err := ioutil.ReadAll(inputReader)
	if err != nil {
		return models.BitriseDataModel{}, nil, err
	}
	decodedBitriseYML, warnings, err := bitrise.ConfigModelFromYAMLBytes(content)
	if err != nil {
		return models.BitriseDataModel{}, nil, fmt.Errorf("Configuration is not valid: %s", err)
	}
	return decodedBitriseYML, warnings, nil
}

func selectBitriseYMLFile(inputReader io.Reader, potentialBitriseYMLFilePath string) (models.BitriseDataModel, error) {
	for {
		const msgBitriseYml = "Enter the path of your bitrise.yml file (you can also drag & drop the file here)"

		var filePath string
		var err error
		if potentialBitriseYMLFilePath != "" {
			filePath, err = goinp.AskForPathFromReaderWithDefault(msgBitriseYml, potentialBitriseYMLFilePath, inputReader)
			if err != nil {
				return models.BitriseDataModel{}, err
			}
		} else {
			filePath, err = goinp.AskForPathFromReader(msgBitriseYml, inputReader)
			if err != nil {
				return models.BitriseDataModel{}, err
			}
		}

		bitriseYMLFile, err := os.Open(filePath)
		defer func() {
			if err := bitriseYMLFile.Close(); err != nil {
				log.Warnf("failed to close file, error: %s", err)
			}
		}()
		if err != nil {
			if !os.IsNotExist(err) {
				return models.BitriseDataModel{}, fmt.Errorf("failed to open file (%s), error: %s", filePath, err)
			}
			log.Warnf("File (%s) does not exist.", filePath)
			continue
		}

		decodedBitriseYML, warnings, err := ParseBitriseYMLFile(bitriseYMLFile)
		if err != nil {
			log.Warnf("Failed to parse bitrise.yml, error: %s", err)
			continue
		} else if len(warnings) > 0 {
			log.Warnf("Parsed bitrise.yml, with warnings:")
			for _, warning := range warnings {
				log.Warnf(warning)
			}
		}
		return decodedBitriseYML, nil
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

	prompt := promptui.Select{
		Label: "Select workflow to run in the first build",
		Items: workflows,
		Templates: &promptui.SelectTemplates{
			Selected: "selected workflow: {{ . }}",
		},
	}

	_, workflow, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return workflow, nil
}

func getBitriseYML(searchDir string, inputReader io.Reader) (models.BitriseDataModel, string, error) {
	potentialBitriseYMLFilePath := filepath.Join(searchDir, bitriseYMLName)
	if exist, err := pathutil.IsPathExists(potentialBitriseYMLFilePath); err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to check if file (%s) exists, error: %s", potentialBitriseYMLFilePath, err)
	} else if exist {
		log.Printf("Found bitrise.yml in current directory.")
	} else {
		potentialBitriseYMLFilePath = ""
	}

	const msg = "What bitrise.yml do you want to upload?"
	const optionRunScanner = "Run the scanner to generate a new bitrise.yml"
	const optionAlreadyExisting = "Use an already existing bitrise.yml"
	options := []string{
		optionRunScanner,
		optionAlreadyExisting,
	}

	prompt := promptui.Select{
		Label: msg,
		Items: options,
		Templates: &promptui.SelectTemplates{
			Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
			Selected: "{{ . }}",
		},
	}

	_, answer, err := prompt.Run()
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to get bitrise.yml, error: %s", err)
	}

	if answer == optionAlreadyExisting {
		bitriseYML, err := selectBitriseYMLFile(inputReader, potentialBitriseYMLFilePath)
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to select bitrise.yml, error: %s", err)
		}

		branch, err := currentBranch(searchDir)
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to get current branch, error: %s", err)
		}

		branchName, err := askBranch(branch.tracking, inputReader)
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to ask for primary branch, error: %s", err)
		}

		return bitriseYML, branchName, nil
	}

	branch, err := checkBranch(searchDir, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to check repository branch: %s", err)
	}

	fmt.Println()

	scanResult, found := scanner.GenerateScanResult(searchDir)
	if !found {
		log.Infof("Projects not found in repository. Select manual configuration.")
		scanResult, err = scanner.ManualConfig()
		if err != nil {
			return models.BitriseDataModel{}, "", fmt.Errorf("failed to get manual configurations, error: %s", err)
		}
	} else {
		var platforms []string
		for scanner := range scanResult.ScannerToOptionRoot {
			platforms = append(platforms, scanner)
		}
		log.Printf("Project(s) found in the repository: %s", colorstring.Green(strings.Join(platforms, ", ")))
	}
	bitriseYML, err := scanner.AskForConfig(scanResult)
	if err != nil {
		return models.BitriseDataModel{}, "", fmt.Errorf("failed to get exact configuration from scanner result, error: %s", err)
	}
	return bitriseYML, branch, nil
}

// BitriseYML ...
func BitriseYML(searchDir string) (models.BitriseDataModel, string, string, error) {
	fmt.Println()
	log.Infof("SETUP BITRISE.YML")
	bitriseYML, branch, err := getBitriseYML(searchDir, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", "", err
	}

	workflow, err := selectWorkflow(bitriseYML, os.Stdin)
	if err != nil {
		return models.BitriseDataModel{}, "", "", fmt.Errorf("failed to select workflow, error: %s", err)
	}
	return bitriseYML, workflow, branch, nil
}
