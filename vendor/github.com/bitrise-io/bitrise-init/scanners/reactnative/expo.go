package reactnative

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitrise-io/bitrise-init/models"
	"github.com/bitrise-io/bitrise-init/scanners/android"
	"github.com/bitrise-io/bitrise-init/scanners/ios"
	"github.com/bitrise-io/bitrise-init/steps"
	"github.com/bitrise-io/bitrise-init/utility"
	envmanModels "github.com/bitrise-io/envman/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	expoConfigName        = "react-native-expo-config"
	expoDefaultConfigName = "default-" + expoConfigName
)

const (
	bareIOSProjectPathInputTitle   = "The iOS project path generated by running 'expo eject' locally"
	bareIOSprojectPathInputSummary = `Will add the Expo Eject Step to the Workflow to generate the native iOS project, so it can be built and archived.
Run 'expo eject' in a local environment to determine this value. This experiment then can be undone by deleting the ios and android directories. See https://docs.expo.io/bare/customizing/ for more details.
For example: './ios/myproject.xcworkspace'.`
)

const (
	iosBundleIDInputTitle   = "iOS bundle identifier"
	iosBundleIDInputSummary = `Key expo/ios/bundleIdentifier not present in 'app.json'.

Will add the Expo Eject Step to the Workflow to generate the native iOS project, so the IPA can be exported.
For your convenience, define it here temporarily. To set this value permanently run 'expo eject' in a local environment and commit 'app.json' changes.
For example: 'com.sample.myapp'.`
	iosBundleIDInputSummaryDefault = `Optional, only needs to be entered if the key expo/ios/bundleIdentifier is not set in 'app.json'.

Will add the Expo Eject Step to the Workflow to generate the native iOS project, so the IPA can be exported.
For your convenience, define it here temporarily. To set this value permanently run 'expo eject' in a local environment and commit 'app.json' changes.
For example: 'com.sample.myapp'.`
	iosBundleIDEnvKey = "EXPO_BARE_IOS_BUNLDE_ID"
)

const (
	androidPackageInputTitle   = "Android package name"
	androidPackageInputSummary = `Key expo/android/package not present in 'app.json'.

Will add the Expo Eject Step to the Workflow to generate the native Android project, so the bundle (AAB) can be built.
For your convenience, define it here temporarily. To set this value permanently run 'expo eject' in a local environment and commit 'app.json' changes.
For example: 'com.sample.myapp'.`
	androidPackageInputSummaryDefault = `Optional, only needs to be entered if the key expo/android/package is not set in 'app.json'.

Will add the Expo Eject Step to the Workflow to generate the native Android project, so the bundle (AAB) can be built.
For your convenience, define it here temporarily. To set this value permanently run 'expo eject' in a local environment and commit 'app.json' changes.
For example: 'com.sample.myapp'.`
	androidPackageEnvKey = "EXPO_BARE_ANDROID_PACKAGE"
)

const (
	iosDevelopmentTeamInputTitle   = "iOS Development team ID"
	iosDevelopmentTeamInputSummary = `The Apple Development Team that the iOS version of the app belongs to. Will be used to override code signing settings. See https://devcenter.bitrise.io/getting-started/getting-started-with-expo-apps/#signing-and-exporting-your-ios-app-for-deployment for more details.

Will add the Expo Eject Step to the Workflow to generate the native iOS project, so it can be built and archived.
Run 'expo eject' in a local environment to determine this value. This experiment then can be undone by deleting the ios and android directories.
For example: '1MZX23ABCD4'.`
	iosDevelopmentTeamEnv = "BITRISE_IOS_DEVELOPMENT_TEAM"
)

const (
	projectRootDirInputTitle   = "Project root directory"
	projectRootDirInputSummary = "The directory of the 'app.json' or 'package.json' file of your React Native project."
)

const (
	schemeInputTitle   = "The iOS native project scheme name"
	schemeInputSummary = `An Xcode scheme defines a collection of targets to build, a configuration to use when building, and a collection of tests to execute. You can change the scheme at any time.

Will add the Expo Eject Step to the Workflow to generate the native iOS project, so it can be built and archived.
Run 'expo eject' in a local environment to determine this value. This experiment then can be undone by deleting the ios and android directories.`
)

const wordirEnv = "WORKDIR"

const (
	expoBareAddIdentiferScriptTitle = "Set bundleIdentifier, packageName for Expo Eject"
	expoAppJSONName                 = "app.json"
)

func expoBareAddIdentifiersScript(appJSONPath, androidEnvKey, iosEnvKey string) string {
	return fmt.Sprintf(`#!/usr/bin/env bash
set -ex

appJson="%s"
tmp="/tmp/app.json"
jq '.expo.android |= if has("package") or env.`+androidEnvKey+` == "" or env.`+androidEnvKey+` == null then . else .package = env.`+androidEnvKey+` end |
.expo.ios |= if has("bundleIdentifier") or env.`+iosEnvKey+` == "" or env.`+iosEnvKey+` == null then . else .bundleIdentifier = env.`+iosEnvKey+` end' <${appJson} >${tmp}
[[ $?==0 ]] && mv -f ${tmp} ${appJson}`, appJSONPath)
}

// expoOptions implements ScannerInterface.Options function for Expo based React Native projects.
func (scanner *Scanner) expoOptions() (models.OptionNode, models.Warnings, error) {
	warnings := models.Warnings{}
	if scanner.expoSettings == nil {
		return models.OptionNode{}, warnings, errors.New("can not generate expo Options, expoSettings is nil")
	}
	if !scanner.expoSettings.isAndroid && !scanner.expoSettings.isIOS {
		return models.OptionNode{}, warnings, errors.New("can not generate expo Option, neither iOS or Android platform detected")
	}

	log.TPrintf("Project name: %v", scanner.expoSettings.name)
	var iosNode *models.OptionNode
	var exportMethodOption *models.OptionNode
	if scanner.expoSettings.isIOS { // ios options
		schemeOption := models.NewOption(ios.SchemeInputTitle, ios.SchemeInputSummary, ios.SchemeInputEnvKey, models.TypeOptionalSelector)

		// predict the ejected project name
		projectName := strings.ToLower(regexp.MustCompile(`(?i:[^a-z0-9])`).ReplaceAllString(scanner.expoSettings.name, ""))
		projectPathOption := models.NewOption(bareIOSProjectPathInputTitle, bareIOSprojectPathInputSummary, ios.ProjectPathInputEnvKey, models.TypeOptionalSelector)
		if projectName != "" {
			projectPathOption.AddOption(filepath.Join("./", "ios", projectName+".xcworkspace"), schemeOption)
		} else {
			projectPathOption.AddOption("", schemeOption)
		}

		if scanner.expoSettings.bundleIdentifierIOS == "" { // bundle ID Option
			iosNode = models.NewOption(iosBundleIDInputTitle, iosBundleIDInputSummary, iosBundleIDEnvKey, models.TypeUserInput)
			iosNode.AddOption("", projectPathOption)
		} else {
			iosNode = projectPathOption
		}

		developmentTeamOption := models.NewOption(iosDevelopmentTeamInputTitle, iosDevelopmentTeamInputSummary, iosDevelopmentTeamEnv, models.TypeUserInput)
		schemeOption.AddOption(projectName, developmentTeamOption)

		exportMethodOption = models.NewOption(ios.IosExportMethodInputTitle, ios.IosExportMethodInputSummary, ios.ExportMethodInputEnvKey, models.TypeSelector)
		developmentTeamOption.AddOption("", exportMethodOption)
	}

	var androidNode *models.OptionNode
	var buildVariantOption *models.OptionNode
	if scanner.expoSettings.isAndroid { // android options
		packageJSONDir := filepath.Dir(scanner.packageJSONPth)
		relPackageJSONDir, err := utility.RelPath(scanner.searchDir, packageJSONDir)
		if err != nil {
			return models.OptionNode{}, warnings, fmt.Errorf("Failed to get relative package.json dir path, error: %s", err)
		}
		if relPackageJSONDir == "." {
			// package.json placed in the search dir, no need to change-dir in the workflows
			relPackageJSONDir = ""
		}

		var projectSettingNode *models.OptionNode
		var moduleOption *models.OptionNode
		if relPackageJSONDir == "" {
			projectSettingNode = models.NewOption(android.ProjectLocationInputTitle, android.ProjectLocationInputSummary, android.ProjectLocationInputEnvKey, models.TypeSelector)

			moduleOption = models.NewOption(android.ModuleInputTitle, android.ModuleInputSummary, android.ModuleInputEnvKey, models.TypeUserInput)
			projectSettingNode.AddOption("./android", moduleOption)
		} else {
			projectSettingNode = models.NewOption(projectRootDirInputTitle, projectRootDirInputSummary, wordirEnv, models.TypeSelector)

			projectLocationOption := models.NewOption(android.ProjectLocationInputTitle, android.ProjectLocationInputSummary, android.ProjectLocationInputEnvKey, models.TypeSelector)
			projectSettingNode.AddOption(relPackageJSONDir, projectLocationOption)

			moduleOption = models.NewOption(android.ModuleInputTitle, android.ModuleInputSummary, android.ModuleInputEnvKey, models.TypeUserInput)
			projectLocationOption.AddOption(filepath.Join(relPackageJSONDir, "android"), moduleOption)
		}

		if scanner.expoSettings.packageNameAndroid == "" {
			androidNode = models.NewOption(androidPackageInputTitle, androidPackageInputSummary, androidPackageEnvKey, models.TypeUserInput)
			androidNode.AddOption("", projectSettingNode)
		} else {
			androidNode = projectSettingNode
		}

		buildVariantOption = models.NewOption(android.VariantInputTitle, android.VariantInputSummary, android.VariantInputEnvKey, models.TypeOptionalUserInput)
		moduleOption.AddOption("app", buildVariantOption)
	}

	rootNode := iosNode
	if iosNode != nil {
		if androidNode != nil {
			for _, exportMethod := range ios.IosExportMethods {
				exportMethodOption.AddOption(exportMethod, androidNode)
			}
		}
	} else {
		rootNode = androidNode
	}

	for _, lastOption := range rootNode.LastChilds() {
		lastOption.ChildOptionMap = map[string]*models.OptionNode{}
		if androidNode != nil {
			// Android buildVariantOption is last
			lastOption.AddConfig("Release", models.NewConfigOption(expoConfigName, nil))
			continue
		}

		// iOS exportMethodOption is last
		for _, exportMethod := range ios.IosExportMethods {
			lastOption.AddConfig(exportMethod, models.NewConfigOption(expoConfigName, nil))
		}
	}

	return *rootNode, warnings, nil
}

// expoConfigs implements ScannerInterface.Configs function for Expo based React Native projects.
func (scanner *Scanner) expoConfigs() (models.BitriseConfigMap, error) {
	configMap := models.BitriseConfigMap{}

	// determine workdir
	packageJSONDir := filepath.Dir(scanner.packageJSONPth)
	relPackageJSONDir, err := utility.RelPath(scanner.searchDir, packageJSONDir)
	if err != nil {
		return models.BitriseConfigMap{}, fmt.Errorf("Failed to get relative package.json dir path, error: %s", err)
	}
	if relPackageJSONDir == "." {
		// package.json placed in the search dir, no need to change-dir in the workflows
		relPackageJSONDir = ""
	}
	log.TPrintf("Working directory: %v", relPackageJSONDir)

	workdirEnvList := []envmanModels.EnvironmentItemModel{}
	if relPackageJSONDir != "" {
		workdirEnvList = append(workdirEnvList, envmanModels.EnvironmentItemModel{workDirInputKey: relPackageJSONDir})
	}

	if !scanner.hasTest {
		// if the project has no test script defined,
		// we can only provide deploy like workflow,
		// so that is going to be the primary workflow

		configBuilder := models.NewDefaultConfigBuilder()
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultPrepareStepList(false)...)

		if scanner.hasYarnLockFile {
			configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.YarnStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
		} else {
			configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
		}

		projectDir := relPackageJSONDir
		if relPackageJSONDir == "" {
			projectDir = "./"
		}

		if !scanner.expoSettings.isAllIdentifierPresent() {
			configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.ScriptSteplistItem(expoBareAddIdentiferScriptTitle,
				envmanModels.EnvironmentItemModel{"content": expoBareAddIdentifiersScript(filepath.Join(projectDir, expoAppJSONName), androidPackageEnvKey, iosBundleIDEnvKey)},
			))
		}

		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.ExpoDetachStepListItem(
			envmanModels.EnvironmentItemModel{"project_path": projectDir},
		))

		// android build
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.InstallMissingAndroidToolsStepListItem(
			envmanModels.EnvironmentItemModel{android.GradlewPathInputKey: "$" + android.ProjectLocationInputEnvKey + "/gradlew"},
		))
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.AndroidBuildStepListItem(
			envmanModels.EnvironmentItemModel{android.ProjectLocationInputKey: "$" + android.ProjectLocationInputEnvKey},
			envmanModels.EnvironmentItemModel{android.ModuleInputKey: "$" + android.ModuleInputEnvKey},
			envmanModels.EnvironmentItemModel{android.VariantInputKey: "$" + android.VariantInputEnvKey},
		))

		// ios build
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.CertificateAndProfileInstallerStepListItem())
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.XcodeArchiveStepListItem(
			envmanModels.EnvironmentItemModel{ios.ProjectPathInputKey: "$" + ios.ProjectPathInputEnvKey},
			envmanModels.EnvironmentItemModel{ios.SchemeInputKey: "$" + ios.SchemeInputEnvKey},
			envmanModels.EnvironmentItemModel{ios.ConfigurationInputKey: "Release"},
			envmanModels.EnvironmentItemModel{ios.ExportMethodInputKey: "$" + ios.ExportMethodInputEnvKey},
			envmanModels.EnvironmentItemModel{"force_team_id": "$BITRISE_IOS_DEVELOPMENT_TEAM"},
		))

		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList(false)...)
		configBuilder.SetWorkflowDescriptionTo(models.PrimaryWorkflowID, deployWorkflowDescription)

		bitriseDataModel, err := configBuilder.Generate(scannerName)
		if err != nil {
			return models.BitriseConfigMap{}, err
		}

		data, err := yaml.Marshal(bitriseDataModel)
		if err != nil {
			return models.BitriseConfigMap{}, err
		}

		configMap[expoConfigName] = string(data)

		return configMap, nil
	}

	// primary workflow
	configBuilder := models.NewDefaultConfigBuilder()
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultPrepareStepList(false)...)
	if scanner.hasYarnLockFile {
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.YarnStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.YarnStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "test"})...))
	} else {
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "test"})...))
	}
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList(false)...)

	// deploy workflow
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultPrepareStepList(false)...)
	if scanner.hasYarnLockFile {
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.YarnStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
	} else {
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.NpmStepListItem(append(workdirEnvList, envmanModels.EnvironmentItemModel{"command": "install"})...))
	}

	projectDir := relPackageJSONDir
	if relPackageJSONDir == "" {
		projectDir = "./"
	}

	if !scanner.expoSettings.isAllIdentifierPresent() {
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.ScriptSteplistItem(expoBareAddIdentiferScriptTitle,
			envmanModels.EnvironmentItemModel{"content": expoBareAddIdentifiersScript(filepath.Join(projectDir, expoAppJSONName), androidPackageEnvKey, iosBundleIDEnvKey)},
		))
	}

	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.ExpoDetachStepListItem(
		envmanModels.EnvironmentItemModel{"project_path": projectDir},
	))

	// android build
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.InstallMissingAndroidToolsStepListItem(
		envmanModels.EnvironmentItemModel{android.GradlewPathInputKey: "$" + android.ProjectLocationInputEnvKey + "/gradlew"},
	))
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.AndroidBuildStepListItem(
		envmanModels.EnvironmentItemModel{android.ProjectLocationInputKey: "$" + android.ProjectLocationInputEnvKey},
		envmanModels.EnvironmentItemModel{android.ModuleInputKey: "$" + android.ModuleInputEnvKey},
		envmanModels.EnvironmentItemModel{android.VariantInputKey: "$" + android.VariantInputEnvKey},
	))

	// ios build
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.CertificateAndProfileInstallerStepListItem())
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.XcodeArchiveStepListItem(
		envmanModels.EnvironmentItemModel{ios.ProjectPathInputKey: "$" + ios.ProjectPathInputEnvKey},
		envmanModels.EnvironmentItemModel{ios.SchemeInputKey: "$" + ios.SchemeInputEnvKey},
		envmanModels.EnvironmentItemModel{ios.ConfigurationInputKey: "Release"},
		envmanModels.EnvironmentItemModel{ios.ExportMethodInputKey: "$" + ios.ExportMethodInputEnvKey},
		envmanModels.EnvironmentItemModel{"force_team_id": "$BITRISE_IOS_DEVELOPMENT_TEAM"},
	))

	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultDeployStepList(false)...)
	configBuilder.SetWorkflowDescriptionTo(models.DeployWorkflowID, deployWorkflowDescription)

	bitriseDataModel, err := configBuilder.Generate(scannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(bitriseDataModel)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	configMap[expoConfigName] = string(data)

	return configMap, nil
}

// expoDefaultOptions implements ScannerInterface.DefaultOptions function for Expo based React Native projects.
func (Scanner) expoDefaultOptions() models.OptionNode {
	// ios options
	rootNode := models.NewOption(bareIOSProjectPathInputTitle, bareIOSprojectPathInputSummary, ios.ProjectPathInputEnvKey, models.TypeUserInput)

	bundleIDOption := models.NewOption(iosBundleIDInputTitle, iosBundleIDInputSummaryDefault, iosBundleIDEnvKey, models.TypeUserInput)
	rootNode.AddOption("", bundleIDOption)

	schemeOption := models.NewOption(schemeInputTitle, schemeInputSummary, ios.SchemeInputEnvKey, models.TypeUserInput)
	bundleIDOption.AddOption("", schemeOption)

	exportMethodOption := models.NewOption(ios.IosExportMethodInputTitle, ios.IosExportMethodInputSummary, ios.ExportMethodInputEnvKey, models.TypeSelector)
	schemeOption.AddOption("", exportMethodOption)

	// android options
	androidPackageOption := models.NewOption(androidPackageInputTitle, androidPackageInputSummaryDefault, androidPackageEnvKey, models.TypeOptionalUserInput)
	for _, exportMethod := range ios.IosExportMethods {
		exportMethodOption.AddOption(exportMethod, androidPackageOption)
	}

	workDirOption := models.NewOption(projectRootDirInputTitle, projectRootDirInputSummary, wordirEnv, models.TypeUserInput)
	androidPackageOption.AddOption("", workDirOption)

	projectLocationOption := models.NewOption(android.ProjectLocationInputTitle, android.ProjectLocationInputSummary, android.ProjectLocationInputEnvKey, models.TypeSelector)
	workDirOption.AddOption("", projectLocationOption)

	moduleOption := models.NewOption(android.ModuleInputTitle, android.ModuleInputSummary, android.ModuleInputEnvKey, models.TypeUserInput)
	projectLocationOption.AddOption("./android", moduleOption)

	buildVariantOption := models.NewOption(android.VariantInputTitle, android.VariantInputSummary, android.VariantInputEnvKey, models.TypeOptionalUserInput)
	moduleOption.AddOption("app", buildVariantOption)

	for _, lastOption := range rootNode.LastChilds() {
		lastOption.ChildOptionMap = map[string]*models.OptionNode{}
		// buildVariantOption is the last Option added
		lastOption.AddConfig("Release", models.NewConfigOption(expoDefaultConfigName, nil))
	}

	return *rootNode
}

// expoDefaultConfigs implements ScannerInterface.DefaultConfigs function for Expo based React Native projects.
func (Scanner) expoDefaultConfigs() (models.BitriseConfigMap, error) {
	configMap := models.BitriseConfigMap{}

	// primary workflow
	configBuilder := models.NewDefaultConfigBuilder()

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultPrepareStepList(false)...)
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.YarnStepListItem(envmanModels.EnvironmentItemModel{workDirInputKey: "$WORKDIR"}, envmanModels.EnvironmentItemModel{"command": "install"}))
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.YarnStepListItem(envmanModels.EnvironmentItemModel{workDirInputKey: "$WORKDIR"}, envmanModels.EnvironmentItemModel{"command": "test"}))
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList(false)...)

	// deploy workflow
	configBuilder.SetWorkflowDescriptionTo(models.DeployWorkflowID, deployWorkflowDescription)
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultPrepareStepList(false)...)
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.YarnStepListItem(envmanModels.EnvironmentItemModel{workDirInputKey: "$WORKDIR"}, envmanModels.EnvironmentItemModel{"command": "install"}))

	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.ScriptSteplistItem(expoBareAddIdentiferScriptTitle,
		envmanModels.EnvironmentItemModel{"content": expoBareAddIdentifiersScript(filepath.Join(".", expoAppJSONName), androidPackageEnvKey, iosBundleIDEnvKey)},
	))

	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.ExpoDetachStepListItem(
		envmanModels.EnvironmentItemModel{"project_path": "$WORKDIR"},
	))

	// android build
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.InstallMissingAndroidToolsStepListItem(
		envmanModels.EnvironmentItemModel{android.GradlewPathInputKey: "$" + android.ProjectLocationInputEnvKey + "/gradlew"},
	))
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.AndroidBuildStepListItem(
		envmanModels.EnvironmentItemModel{android.ProjectLocationInputKey: "$" + android.ProjectLocationInputEnvKey},
		envmanModels.EnvironmentItemModel{android.ModuleInputKey: "$" + android.ModuleInputEnvKey},
		envmanModels.EnvironmentItemModel{android.VariantInputKey: "$" + android.VariantInputEnvKey},
	))

	// ios build
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.CertificateAndProfileInstallerStepListItem())
	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.XcodeArchiveStepListItem(
		envmanModels.EnvironmentItemModel{ios.ProjectPathInputKey: "$" + ios.ProjectPathInputEnvKey},
		envmanModels.EnvironmentItemModel{ios.SchemeInputKey: "$" + ios.SchemeInputEnvKey},
		envmanModels.EnvironmentItemModel{ios.ExportMethodInputKey: "$" + ios.ExportMethodInputEnvKey},
		envmanModels.EnvironmentItemModel{ios.ConfigurationInputKey: "Release"},
	))

	configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultDeployStepList(false)...)

	bitriseDataModel, err := configBuilder.Generate(scannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(bitriseDataModel)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	configMap[expoDefaultConfigName] = string(data)

	return configMap, nil
}
