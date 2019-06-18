package phases

import (
	"fmt"
	"path/filepath"
	"sort"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/codesigndoc"
	"github.com/bitrise-io/codesigndoc/models"
	"github.com/bitrise-io/codesigndoc/xcode"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/sliceutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-io/xcode-project/xcscheme"
	"github.com/bitrise-io/xcode-project/xcworkspace"
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
	if isIOSCodesign(bitriseYML.ProjectType) {
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

func iosCodesign(bitriseYML bitriseModels.BitriseDataModel, searchDir string) (CodesignResultsIOS, error) {
	appEnvToValue, err := evniromentsToMap(bitriseYML.App.Environments)
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	projectPath, pathOk := appEnvToValue["BITRISE_PROJECT_PATH"]
	scheme, schemeOk := appEnvToValue["BITRISE_SCHEME"]

	if !(pathOk && schemeOk) {
		log.Debugf("could not find Xcode project path and scheme in bitrise.yml")

		projectPath, err = askXcodeProjectPath()
		if err != nil {
			return CodesignResultsIOS{}, fmt.Errorf("failed to get Xcode project path, error: %s", err)
		}

		scheme, err = askXcodeProjectScheme(projectPath)
		if err != nil {
			return CodesignResultsIOS{}, fmt.Errorf("failed to get Xcode scheme, error: %s", err)
		}
	} else {
		log.Debugf("Found Xcode project path (%s), scheme (%s) in bitrise.yml.", projectPath, scheme)
	}

	projectPathAbs, err := filepath.Abs(projectPath)
	if err != nil {
		return CodesignResultsIOS{}, err
	}

	archivePath, err := codesigndoc.BuildXcodeArchive(xcode.CommandModel{
		ProjectFilePath: projectPathAbs,
		Scheme:          scheme,
	}, nil)
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

func askXcodeProjectPath() (string, error) {
	for {
		log.Infof("Provide the project file manually")
		askText := `Please drag-and-drop your Xcode Project (` + colorstring.Green(".xcodeproj") + `) or Workspace (` + colorstring.Green(".xcworkspace") + `) file, 
the one you usually open in Xcode, then hit Enter.
(Note: if you have a Workspace file you should most likely use that)`
		path, err := goinp.AskForPath(askText)
		if err != nil {
			return "", fmt.Errorf("failed to read input: %s", err)
		}

		validProject := true

		exists, err := pathutil.IsDirExists(path)
		if err != nil {
			return "", fmt.Errorf("failed to check if project exists, error: %s", err)
		}
		if !exists {
			validProject = false
			log.Warnf("Project directory does not exist.")
		}

		if validProject && !(xcodeproj.IsXcodeProj(path) || xcworkspace.IsWorkspace(path)) {
			validProject = false
			log.Warnf("Directory is not an Xcode project or workspace.")
		}

		if !validProject {
			retry, err := goinp.AskForBoolWithDefault("Input Xcode project or workspace path again?", true)
			if err != nil {
				return "", err
			}

			if retry {
				continue
			}
		}

		return path, nil
	}
}

func askXcodeProjectScheme(path string) (string, error) {
	var schemes []xcscheme.Scheme

	if xcodeproj.IsXcodeProj(path) {
		project, err := xcodeproj.Open(path)
		if err != nil {
			return "", err
		}

		schemes, err = project.Schemes()
		if err != nil {
			return "", err
		}
	} else if xcworkspace.IsWorkspace(path) {
		workspace, err := xcworkspace.Open(path)
		if err != nil {
			return "", err
		}

		projectToScheme, err := workspace.Schemes()
		if err != nil {
			return "", err
		}

		// Sort schemes by project
		var keys []string
		for k := range projectToScheme {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			schemes = append(schemes, projectToScheme[key]...)
		}
	}

	var schemeNames []string
	for _, scheme := range schemes {
		schemeNames = append(schemeNames, scheme.Name)
	}

	if len(schemeNames) == 0 {
		return "", fmt.Errorf("no schemes found in project")
	}

	selectedScheme, err := goinp.SelectFromStringsWithDefault("Select scheme:", 1, schemeNames)
	if err != nil {
		return "", err
	}

	return selectedScheme, nil
}
