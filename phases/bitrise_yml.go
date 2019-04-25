package phases

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

const bitriseYMLName = "bitrise.yml"

func currentBranch() (string, error) {
	out, err := command.New("git", "symbolic-ref", "HEAD").RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, out)
	}
	return strings.TrimPrefix(out, "refs/heads/"), nil
}

// BitriseYML ...
func BitriseYML() (string, string, error) {
	log.Infof("Setup " + bitriseYMLName)
	fmt.Println()

	var (
		err                           error
		bitriseYMLPath, firstWorkflow string
	)

	bitriseYMLFile, openErr := os.Open(bitriseYMLName)
	if openErr != nil && !os.IsNotExist(openErr) {
		return "", "", fmt.Errorf("failed to open "+bitriseYMLName+", error: %s", openErr)
	}

	if bitriseYMLFile == nil {
		log.Warnf("No " + bitriseYMLName + " found in the current working directory.")
		fmt.Println()
		const (
			methodTitle     = "Select method to have a " + bitriseYMLName + ""
			methodInputPath = "Input path manually"
			methodScan      = "Generate one by scanning my project"
		)
		(&option{
			title:        methodTitle,
			valueOptions: []string{methodInputPath, methodScan},
			action: func(answer string) *option {
				const (
					inputTitle = "Enter the path of your " + bitriseYMLName + " file (you can also drag & drop the file here)"
				)
				switch answer {
				case methodInputPath:
					return &option{
						title: inputTitle,
						action: func(answer string) *option {
							bitriseYMLPath = answer
							return nil
						}}
				case methodScan:
					branch, branchReadErr := currentBranch()
					if branchReadErr != nil {
						err = branchReadErr
						return nil
					}

					log.Warnf("The current branch is: %s. Do you want to run the scanner for this branch?", branch)
					log.Warnf("At this point you can checkout a different branch if you want to.")
					fmt.Println()

					var scanOption *option
					scanOption = &option{
						title: "Hit enter if you have done the checkout or want to use this branch",
						action: func(answer string) *option {
							newBranch, branchReadErr := currentBranch()
							if branchReadErr != nil {
								err = branchReadErr
								return nil
							}
							if newBranch == branch {
								if err = command.New("bitrise", ":init").SetStderr(os.Stderr).SetStdin(os.Stdin).SetStdout(os.Stdout).Run(); err != nil {
									return nil
								}
								fmt.Println()
								bitriseYMLPath = bitriseYMLName
								return nil
							}

							branch = newBranch

							log.Warnf("Checked out a different branch: %s. Do you want to run the scanner for this branch?", branch)
							log.Warnf("At this point you can still checkout a different branch if you want to.")
							fmt.Println()

							return scanOption
						}}
					return scanOption
				}
				return nil
			}}).run()
	} else {
		bitriseYMLPath = bitriseYMLName
	}

	if bitriseYMLFile == nil {
		bitriseYMLFile, openErr = os.Open(bitriseYMLPath)
		if openErr != nil {
			return "", "", fmt.Errorf("failed to open "+bitriseYMLPath+", error: %s", openErr)
		}
	}

	var bitriseYML models.BitriseDataModel
	if parseErr := yaml.NewDecoder(bitriseYMLFile).Decode(&bitriseYML); parseErr != nil {
		return "", "", fmt.Errorf("failed to parse "+bitriseYMLPath+", error: %s", parseErr)
	}

	var workflows []string
	for workflow := range bitriseYML.Workflows {
		workflows = append(workflows, workflow)
	}

	(&option{
		title:        "Which workflow do you want to run in the first build?",
		valueOptions: workflows,
		action: func(answer string) *option {
			firstWorkflow = answer
			return nil
		}}).run()

	return bitriseYMLPath, firstWorkflow, err
}
