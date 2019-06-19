package phases

import (
	"fmt"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/goinp/goinp"
	"runtime"
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
		log.Warnf("Unsupported project type (%s) for exporting codesigning files.", bitriseYML.ProjectType)
		log.Warnf("Supported project types for exporting codesigning files: 'ios', 'android'.")
		return CodesignResult{}, nil
	}

	log.Donef("Project type: %s", bitriseYML.ProjectType)
	fmt.Println()

	var result CodesignResult
	if runtime.GOOS == "darwin" && isIOSCodesign(bitriseYML.ProjectType) {
		uploadIOS, err := goinp.AskForBoolWithDefault("Do you want to export and upload iOS codesigning files?", true)
		if err != nil {
			return CodesignResult{}, err
		}

		if uploadIOS {
			log.Infof("Exporting iOS codesigning files.")

			var err error
			for { // The retry is needed as codesign flow contains questions which can not be retried
				result.IOS, err = iosCodesign(bitriseYML, searchDir)
				if err != nil {
					log.Warnf("Failed to export iOS codesigning files, error: %s", err)
					isRetry, err := goinp.AskForBoolWithDefault("Retry exporting iOS codesigning files?", true)
					if err != nil {
						return CodesignResult{}, err
					}
					if isRetry {
						continue
					}
				}
				break
			}
		}
	}

	if isAndroidCodesign(bitriseYML.ProjectType) {
		uploadAndroid, err := goinp.AskForBoolWithDefault("Do you want to upload an Android keystore file?", true)
		if err != nil {
			return CodesignResult{}, err
		}

		if uploadAndroid {
			(&option{
				title: "Enter key store path",
				action: func(answer string) *option {
					result.Android.KeystorePath = answer
					return nil
				}}).run()

			(&option{
				title:  "Enter key store password",
				secret: true,
				action: func(answer string) *option {
					result.Android.Password = answer
					return nil
				}}).run()

			(&option{
				title: "Enter key alias",
				action: func(answer string) *option {
					result.Android.Alias = answer
					return nil
				}}).run()

			(&option{
				title:  "Enter key password",
				secret: true,
				action: func(answer string) *option {
					result.Android.KeyPassword = answer
					return nil
				}}).run()
		}
	}

	return result, nil
}
