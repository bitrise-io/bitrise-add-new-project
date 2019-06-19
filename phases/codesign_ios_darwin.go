package phases

import (
	"fmt"
	"path/filepath"
	"sort"

	bitriseModels "github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/codesigndoc/codesigndoc"
	"github.com/bitrise-io/codesigndoc/xcode"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/goinp/goinp"
	"github.com/bitrise-io/xcode-project/xcodeproj"
	"github.com/bitrise-io/xcode-project/xcscheme"
	"github.com/bitrise-io/xcode-project/xcworkspace"
)

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
