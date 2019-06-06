package phases

import (
	"fmt"
	"path/filepath"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/codesigndoc"
	"github.com/bitrise-io/codesigndoc/models"
	"github.com/bitrise-io/codesigndoc/xcode"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/goinp/goinp"
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

// Project types: "web", "macos", "other", "": is not suppored

var codesignBothPlatforms = []string{"xamarin", "flutter", "cordova", "ionic", "react-native"}

func isIOSCodesign(projectType string) bool {
	return projectType == "ios" || sliceutil.IsStringInSlice(projectType, codesignBothPlatforms)
}

func isAndroidCodesign(projectType string) bool {
	return projectType == "android" || sliceutil.IsStringInSlice(projectType, codesignBothPlatforms)
}

// AutoCodesign ...
func AutoCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResult, error) {
	log.Donef("Project type: %s", bitriseYML.ProjectType)
	fmt.Println()

	var result CodesignResult

	isExport, err := goinp.AskForBool("Do you want to export and upload codesigning files?")
	if err != nil {
		return CodesignResult{}, err
	}
	if isExport {
		if isIOSCodesign(bitriseYML.ProjectType) {
			log.Infof("Exporting iOS codesigning files")

			var err error
			for { // The retry is needed as codesign flow contains questions which can not be retried
				result.IOS, err = iosCodesign(bitriseYML, searchDir)
				if err != nil {
					log.Warnf("Failed to export iOS codesigning files, error: %s", err)
					isRetry, err := goinp.AskForBoolWithDefault("Retry exporting iOS codesigning files?", true)
					if err != nil {
						return CodesignResult{}, err
					}
					if !isRetry {
						break
					}
				}
				break
			}
		}

		if isAndroidCodesign(bitriseYML.ProjectType) {
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

	return result, err
}

func evniromentsToMap(envs []envmanModels.EnvironmentItemModel) (map[string]string, error) {
	nameToValue := map[string]string{}

	for _, env := range envs {
		key, value, err := env.GetKeyValuePair()
		if err != nil {
			return nil, err
		}
		nameToValue[key] = value
	}

	return nameToValue, nil
}

func iosCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResultsIOS, error) {
	var xcodeProjects []xcode.CommandModel

	appEnvToValue, err := evniromentsToMap(bitriseYML.App.Environments)
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	projectPath, pathOk := appEnvToValue["BITRISE_PROJECT_PATH"]
	scheme, schemeOk := appEnvToValue["BITRISE_SCHEME"]

	if !(pathOk && schemeOk) {
		return CodesignResultsIOS{}, fmt.Errorf("could not find Xcode project path")
	}

	projectPathAbs, err := filepath.Abs(projectPath)
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	xcodeProjects = append(xcodeProjects, xcode.CommandModel{
		ProjectFilePath: projectPathAbs,
		Scheme:          scheme,
	})
	log.Debugf("Xcode projects: %s", xcodeProjects)

	// ToDo: add list of found Xcode projects to choose from

	archivePath, _, err := codesigndoc.GenerateXCodeArchive(xcodeProjects[0])
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	certificates, profiles, err := codesigndoc.CodesigningFilesForXCodeProject(archivePath, false, false)
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	log.Debugf("Certificates: %s \nProfiles: %s", certificates, profiles)
	return CodesignResultsIOS{
		certificates:         certificates,
		provisioningProfiles: profiles,
	}, nil
}
