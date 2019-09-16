package phases

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/manifoldco/promptui"
)

// CodesignResultsIOS ...
type CodesignResultsIOS struct {
	certificates         models.Certificates
	provisioningProfiles []models.ProvisioningProfile
}

// CodesignResultAndroid ...
type CodesignResultAndroid struct {
	KeystorePath, Password, Alias, KeyPassword string
}

// CodesignResult ...
type CodesignResult struct {
	Android CodesignResultAndroid
	IOS     CodesignResultsIOS
}

// Project types "web", "macos" are unsuppored as do not have ios and android native projects
// Project types "xamarin", "flutter", "cordova", "ionic", "react-native" are unsupported, due to ucertainty that native (Xcode or Android) project is present

var unknownPlatforms = []string{"", "other"}

func isIOSCodesign(projectType string) bool {
	return projectType == "ios" ||
		sliceutil.IsStringInSlice(projectType, unknownPlatforms)
}

func isAndroidCodesign(projectType string) bool {
	return projectType == "android" ||
		sliceutil.IsStringInSlice(projectType, unknownPlatforms)
}

// AutoCodesign ...
func AutoCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResult, error) {
	if !isIOSCodesign(bitriseYML.ProjectType) && !isAndroidCodesign(bitriseYML.ProjectType) {
		return CodesignResult{}, nil
	}

	fmt.Println()
	log.Infof("CODESIGNING")

	log.Debugf("Project type: %s", bitriseYML.ProjectType)

	const (
		answerYes = "Yes"
		answerNo  = "No"
	)

	var result CodesignResult
	if runtime.GOOS == "darwin" && isIOSCodesign(bitriseYML.ProjectType) {
		prompt := promptui.Select{
			Label: "Do you want to export and upload iOS codesigning files?",
			Items: []string{answerYes, answerNo},
			Templates: &promptui.SelectTemplates{
				Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
				Selected: "Export and upload iOS codesigning files: {{ . | green }}",
			},
		}
		_, uploadIOS, err := prompt.Run()
		if err != nil {
			return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
		}

		if uploadIOS == answerYes {
			log.Debugf("Exporting iOS codesigning files.")

			var err error
			for { // The retry is needed as codesign flow contains questions which can not be retried
				result.IOS, err = iosCodesign(bitriseYML, searchDir)
				if err != nil {
					log.Warnf("Failed to export iOS codesigning files, error: %s", err)
					prompt := promptui.Select{
						Label: "Retry exporting iOS codesigning files?",
						Items: []string{answerYes, answerNo},
						Templates: &promptui.SelectTemplates{
							Selected: "",
						},
					}
					_, isRetry, err := prompt.Run()
					if err != nil {
						return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
					}
					if isRetry == answerYes {
						continue
					}
				}
				break
			}
		}
	}

	if isAndroidCodesign(bitriseYML.ProjectType) {
		for {
			prmpt := promptui.Select{
				Label: "Do you want to upload an Android keystore file",
				Items: []string{answerYes, answerNo},
				Templates: &promptui.SelectTemplates{
					Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
					Selected: "Upload Android keystore file: {{ . | green }}",
				},
			}
			_, uploadAndroid, err := prmpt.Run()
			if err != nil {
				return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
			}

			if uploadAndroid == answerNo {
				break
			}

			result.Android, err = getAndroidKeystoreSettings()
			if err != nil {
				log.Errorf("%s", err)
				continue
			}

			break
		}
	}

	return result, nil
}

func getAndroidKeystoreSettings() (CodesignResultAndroid, error) {
	prompt := promptui.Prompt{
		Label: "Enter keystore path",
		Templates: &promptui.PromptTemplates{
			Success: "Keystore path: {{ . | green }}",
		},
	}

	var absKeystorePath string
	{
		keystorePath, err := prompt.Run()
		if err != nil {
			return CodesignResultAndroid{}, fmt.Errorf("scan user input: %s", err)
		}

		absKeystorePath, err = pathutil.AbsPath(keystorePath)
		if err != nil {
			return CodesignResultAndroid{}, fmt.Errorf("failed to get absolute keystore path, error: %s", err)
		}
		if _, err := os.Stat(absKeystorePath); os.IsNotExist(err) {
			return CodesignResultAndroid{}, fmt.Errorf("keystore file does not exist, error: %s", err)
		}
	}

	prompt = promptui.Prompt{
		Label: "Enter key store password",
		Mask:  '*',
		Templates: &promptui.PromptTemplates{
			Success: "Keystore password: [REDACTED]",
		},
	}
	keystorePassword, err := prompt.Run()
	if err != nil {
		return CodesignResultAndroid{}, fmt.Errorf("scan user input: %s", err)
	}

	prompt = promptui.Prompt{
		Label: "Enter key alias",
		Templates: &promptui.PromptTemplates{
			Success: "Key alias: {{ . | green }}",
		},
	}
	alias, err := prompt.Run()
	if err != nil {
		return CodesignResultAndroid{}, fmt.Errorf("scan user input: %s", err)
	}

	prompt = promptui.Prompt{
		Label: "Enter key password",
		Mask:  '*',
		Templates: &promptui.PromptTemplates{
			Success: "Key password: [REDACTED]",
		},
	}

	keyPassword, err := prompt.Run()
	if err != nil {
		return CodesignResultAndroid{}, fmt.Errorf("scan user input: %s", err)
	}

	keystoreSettings := CodesignResultAndroid{
		KeystorePath: absKeystorePath,
		Password:     keystorePassword,
		Alias:        alias,
		KeyPassword:  keyPassword,
	}

	if err := validateAndroidCodesignParams(keystoreSettings); err != nil {
		return CodesignResultAndroid{}, fmt.Errorf("Invalid keystore parameters, error: %s", err)
	}

	return keystoreSettings, nil
}

func validateAndroidCodesignParams(codesign CodesignResultAndroid) error {
	params := []string{
		"-certreq",
		"-v",

		"-keystore",
		codesign.KeystorePath,
		"-storepass",
		codesign.Password,

		"-alias",
		codesign.Alias,
		"-keypass",
		codesign.KeyPassword,

		"-J-Dfile.encoding=utf-8",
		"-J-Duser.language=en-US",
	}

	out, err := command.New("keytool", params...).RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return errors.New(out)
		}
		return fmt.Errorf("failed to run keytool command, error: %s", err)
	}
	if out == "" {
		return fmt.Errorf("failed to read keystore, maybe alias (%s) or keystore password is not correct", codesign.Alias)
	}
	return nil
}
