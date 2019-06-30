package phases

import (
	"fmt"

	"runtime"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/models"
	"github.com/bitrise-io/go-utils/log"
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
	fmt.Println()
	log.Infof("CODESIGNING")

	if !isIOSCodesign(bitriseYML.ProjectType) && !isAndroidCodesign(bitriseYML.ProjectType) {
		log.Warnf("Unsupported project type (%s) for exporting codesigning files.", bitriseYML.ProjectType)
		log.Warnf("Supported project types for exporting codesigning files: 'ios', 'android'.")
		return CodesignResult{}, nil
	}

	log.Debugf("Project type: %s", bitriseYML.ProjectType)

	const (
		answerYes = "yes"
		answerNo  = "no"
	)

	var result CodesignResult
	if runtime.GOOS == "darwin" && isIOSCodesign(bitriseYML.ProjectType) {
		prompt := promptui.Select{
			Label: "Do you want to export and upload iOS codesigning files?",
			Items: []string{answerYes, answerNo},
			Templates: &promptui.SelectTemplates{
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
		prmpt := promptui.Select{
			Label: "Do you want to upload an Android keystore file?",
			Items: []string{answerYes, answerNo},
			Templates: &promptui.SelectTemplates{
				Selected: "Upload Android keystore file: {{ . | green }}",
			},
		}
		_, uploadAndroid, err := prmpt.Run()
		if err != nil {
			return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
		}

		if uploadAndroid == answerYes {
			prompt := promptui.Prompt{
				Label: "Enter keystore path",
				Templates: &promptui.PromptTemplates{
					Success: "Keystore path: {{ . | green }}",
				},
			}

			result.Android.KeystorePath, err = prompt.Run()
			if err != nil {
				return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
			}

			prompt = promptui.Prompt{
				Label: "Enter key store password",
				Mask:  '*',
				Templates: &promptui.PromptTemplates{
					Success: "keystore password: [REDACTED]",
				},
			}
			result.Android.Password, err = prompt.Run()
			if err != nil {
				return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
			}

			prompt = promptui.Prompt{
				Label: "Enter key alias",
				Templates: &promptui.PromptTemplates{
					Success: "key alias: {{ . | green }}",
				},
			}
			result.Android.Alias, err = prompt.Run()
			if err != nil {
				return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
			}

			prompt = promptui.Prompt{
				Label: "Enter key password",
				Mask:  '*',
			}

			result.Android.KeyPassword, err = prompt.Run()
			if err != nil {
				return CodesignResult{}, fmt.Errorf("scan user input: %s", err)
			}
		}
	}

	return result, nil
}
