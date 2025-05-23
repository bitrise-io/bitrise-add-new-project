package cordova

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/bitrise-io/bitrise-init/models"
	"github.com/bitrise-io/bitrise-init/scanners/android"
	"github.com/bitrise-io/bitrise-init/scanners/ios"
	"github.com/bitrise-io/bitrise-init/scanners/java"
	"github.com/bitrise-io/bitrise-init/scanners/nodejs"
	"github.com/bitrise-io/bitrise-init/steps"
	"github.com/bitrise-io/bitrise-init/utility"
	envmanModels "github.com/bitrise-io/envman/v2/models"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
)

// ScannerName ...
const ScannerName = "cordova"

const (
	configName        = "cordova-config"
	defaultConfigName = "default-cordova-config"
)

// Step Inputs
const (
	workDirInputKey     = "workdir"
	workDirInputTitle   = "Directory of the Cordova config.xml file"
	workDirInputEnvKey  = "CORDOVA_WORK_DIR"
	workDirInputSummary = "The working directory of your Cordova project is where you store your config.xml file. In your Workflows, you can specify paths relative to this path. You can change this at any time."
)

const (
	platformInputKey     = "platform"
	platformInputTitle   = "The platform to use in cordova-cli commands"
	platformInputEnvKey  = "CORDOVA_PLATFORM"
	platformInputSummary = "The target platform for your build, stored as an Environment Variable. Your options are iOS, Android, or both. You can change this in your Env Vars at any time."
)

const (
	targetInputKey = "target"
	targetEmulator = "emulator"
)

//------------------
// ScannerInterface
//------------------

// Scanner ...
type Scanner struct {
	cordovaConfigPth    string
	relCordovaConfigDir string
	searchDir           string
	hasKarmaJasmineTest bool
	hasJasmineTest      bool
}

// NewScanner ...
func NewScanner() *Scanner {
	return &Scanner{}
}

// Name ...
func (Scanner) Name() string {
	return ScannerName
}

// DetectPlatform ...
func (scanner *Scanner) DetectPlatform(searchDir string) (bool, error) {
	fileList, err := pathutil.ListPathInDirSortedByComponents(searchDir, true)
	if err != nil {
		return false, fmt.Errorf("failed to search for files in (%s), error: %s", searchDir, err)
	}

	// Search for config.xml file
	log.TInfof("Searching for config.xml file")

	configXMLPth, err := FilterRootConfigXMLFile(fileList)
	if err != nil {
		return false, fmt.Errorf("failed to search for config.xml file, error: %s", err)
	}

	log.TPrintf("config.xml: %s", configXMLPth)

	if configXMLPth == "" {
		log.TPrintf("platform not detected")
		return false, nil
	}

	widget, err := ParseConfigXML(configXMLPth)
	if err != nil {
		log.TPrintf("can not parse config.xml as a Cordova widget, error: %s", err)
		log.TPrintf("platform not detected")
		return false, nil
	}

	// ensure it is a cordova widget
	if !strings.Contains(widget.XMLNSCDV, "cordova.apache.org") {
		log.TPrintf("config.xml propert: xmlns:cdv does not contain cordova.apache.org")
		log.TPrintf("platform not detected")
		return false, nil
	}

	// ensure it is not an ionic project
	projectBaseDir := filepath.Dir(configXMLPth)

	if exist, err := pathutil.IsPathExists(filepath.Join(projectBaseDir, "ionic.project")); err != nil {
		return false, fmt.Errorf("failed to check if project is an ionic project, error: %s", err)
	} else if exist {
		log.TPrintf("ionic.project file found seems to be an ionic project")
		return false, nil
	}

	if exist, err := pathutil.IsPathExists(filepath.Join(projectBaseDir, "ionic.config.json")); err != nil {
		return false, fmt.Errorf("failed to check if project is an ionic project, error: %s", err)
	} else if exist {
		log.TPrintf("ionic.config.json file found seems to be an ionic project")
		return false, nil
	}

	log.TSuccessf("Platform detected")

	scanner.cordovaConfigPth = configXMLPth
	scanner.searchDir = searchDir

	return true, nil
}

// ExcludedScannerNames ...
func (*Scanner) ExcludedScannerNames() []string {
	return []string{
		string(ios.XcodeProjectTypeIOS),
		string(ios.XcodeProjectTypeMacOS),
		android.ScannerName,
		nodejs.ScannerName,
		java.ProjectType,
	}
}

// Options ...
func (scanner *Scanner) Options() (models.OptionNode, models.Warnings, models.Icons, error) {
	warnings := models.Warnings{}
	projectRootDir := filepath.Dir(scanner.cordovaConfigPth)

	packagesJSONPth := filepath.Join(projectRootDir, "package.json")
	packages, err := utility.ParsePackagesJSON(packagesJSONPth)
	if err != nil {
		return models.OptionNode{}, warnings, nil, err
	}

	// Search for karma/jasmine tests
	log.TPrintf("Searching for karma/jasmine test")

	karmaTestDetected := false

	karmaJasmineDependencyFound := false
	for dependency := range packages.Dependencies {
		if strings.Contains(dependency, "karma-jasmine") {
			karmaJasmineDependencyFound = true
		}
	}
	if !karmaJasmineDependencyFound {
		for dependency := range packages.DevDependencies {
			if strings.Contains(dependency, "karma-jasmine") {
				karmaJasmineDependencyFound = true
			}
		}
	}
	log.TPrintf("karma-jasmine dependency found: %v", karmaJasmineDependencyFound)

	if karmaJasmineDependencyFound {
		karmaConfigJSONPth := filepath.Join(projectRootDir, "karma.conf.js")
		if exist, err := pathutil.IsPathExists(karmaConfigJSONPth); err != nil {
			return models.OptionNode{}, warnings, nil, err
		} else if exist {
			karmaTestDetected = true
		}
	}
	log.TPrintf("karma.conf.js found: %v", karmaTestDetected)

	scanner.hasKarmaJasmineTest = karmaTestDetected
	// ---

	// Search for jasmine tests
	jasminTestDetected := false

	if !karmaTestDetected {
		log.TPrintf("Searching for jasmine test")

		jasmineDependencyFound := false
		for dependency := range packages.Dependencies {
			if strings.Contains(dependency, "jasmine") {
				jasmineDependencyFound = true
				break
			}
		}
		if !jasmineDependencyFound {
			for dependency := range packages.DevDependencies {
				if strings.Contains(dependency, "jasmine") {
					jasmineDependencyFound = true
					break
				}
			}
		}
		log.TPrintf("jasmine dependency found: %v", jasmineDependencyFound)

		if jasmineDependencyFound {
			jasmineConfigJSONPth := filepath.Join(projectRootDir, "spec", "support", "jasmine.json")
			if exist, err := pathutil.IsPathExists(jasmineConfigJSONPth); err != nil {
				return models.OptionNode{}, warnings, nil, err
			} else if exist {
				jasminTestDetected = true
			}
		}

		log.TPrintf("jasmine.json found: %v", jasminTestDetected)

		scanner.hasJasmineTest = jasminTestDetected
	}
	// ---

	// Get relative config.xml dir
	cordovaConfigDir := filepath.Dir(scanner.cordovaConfigPth)
	relCordovaConfigDir, err := utility.RelPath(scanner.searchDir, cordovaConfigDir)
	if err != nil {
		return models.OptionNode{}, warnings, nil, fmt.Errorf("Failed to get relative config.xml dir path, error: %s", err)
	}
	if relCordovaConfigDir == "." {
		// config.xml placed in the search dir, no need to change-dir in the workflows
		relCordovaConfigDir = ""
	}
	scanner.relCordovaConfigDir = relCordovaConfigDir
	// ---

	// Options
	var rootOption *models.OptionNode

	platforms := []string{"ios", "android", "ios,android"}

	if relCordovaConfigDir != "" {
		rootOption = models.NewOption(workDirInputTitle, workDirInputSummary, workDirInputEnvKey, models.TypeSelector)

		platformTypeOption := models.NewOption(platformInputTitle, platformInputSummary, platformInputEnvKey, models.TypeSelector)
		rootOption.AddOption(relCordovaConfigDir, platformTypeOption)

		for _, platform := range platforms {
			configOption := models.NewConfigOption(configName, nil)
			platformTypeOption.AddConfig(platform, configOption)
		}
	} else {
		rootOption = models.NewOption(platformInputTitle, platformInputSummary, platformInputEnvKey, models.TypeSelector)

		for _, platform := range platforms {
			configOption := models.NewConfigOption(configName, nil)
			rootOption.AddConfig(platform, configOption)
		}
	}
	// ---

	return *rootOption, warnings, nil, nil
}

// DefaultOptions ...
func (*Scanner) DefaultOptions() models.OptionNode {
	workDirOption := models.NewOption(workDirInputTitle, workDirInputSummary, workDirInputEnvKey, models.TypeUserInput)

	platformTypeOption := models.NewOption(platformInputTitle, platformInputSummary, platformInputEnvKey, models.TypeSelector)
	workDirOption.AddOption(models.UserInputOptionDefaultValue, platformTypeOption)

	platforms := []string{
		"ios",
		"android",
		"ios,android",
	}
	for _, platform := range platforms {
		configOption := models.NewConfigOption(defaultConfigName, nil)
		platformTypeOption.AddConfig(platform, configOption)
	}

	return *workDirOption
}

func (scanner *Scanner) Configs(sshKeyActivation models.SSHKeyActivation) (models.BitriseConfigMap, error) {
	configBuilder := models.NewDefaultConfigBuilder()
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultPrepareStepList(steps.PrepareListParams{
		SSHKeyActivation: sshKeyActivation,
	})...)

	workdirEnvList := []envmanModels.EnvironmentItemModel{}
	workdir := ""
	if scanner.relCordovaConfigDir != "" {
		workdir = "$" + workDirInputEnvKey
		workdirEnvList = append(workdirEnvList, envmanModels.EnvironmentItemModel{workDirInputKey: workdir})
	}

	if scanner.hasJasmineTest || scanner.hasKarmaJasmineTest {
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.RestoreNPMCache())
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem("install", workdir))

		// CI
		if scanner.hasKarmaJasmineTest {
			configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.KarmaJasmineTestRunnerStepListItem(workdirEnvList...))
		} else if scanner.hasJasmineTest {
			configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.JasmineTestRunnerStepListItem(workdirEnvList...))
		}
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.SaveNPMCache())
		configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList()...)

		// CD
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultPrepareStepList(steps.PrepareListParams{
			SSHKeyActivation: sshKeyActivation,
		})...)
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.CertificateAndProfileInstallerStepListItem())

		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.NpmStepListItem("install", workdir))

		if scanner.hasKarmaJasmineTest {
			configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.KarmaJasmineTestRunnerStepListItem(workdirEnvList...))
		} else if scanner.hasJasmineTest {
			configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.JasmineTestRunnerStepListItem(workdirEnvList...))
		}

		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.GenerateCordovaBuildConfigStepListItem())

		cordovaArchiveEnvs := []envmanModels.EnvironmentItemModel{
			{platformInputKey: "$" + platformInputEnvKey},
			{targetInputKey: targetEmulator},
		}
		if scanner.relCordovaConfigDir != "" {
			cordovaArchiveEnvs = append(cordovaArchiveEnvs, envmanModels.EnvironmentItemModel{workDirInputKey: "$" + workDirInputEnvKey})
		}
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.CordovaArchiveStepListItem(cordovaArchiveEnvs...))
		configBuilder.AppendStepListItemsTo(models.DeployWorkflowID, steps.DefaultDeployStepList()...)

		config, err := configBuilder.Generate(ScannerName)
		if err != nil {
			return models.BitriseConfigMap{}, err
		}

		data, err := yaml.Marshal(config)
		if err != nil {
			return models.BitriseConfigMap{}, err
		}

		return models.BitriseConfigMap{
			configName: string(data),
		}, nil
	}

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.CertificateAndProfileInstallerStepListItem())
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.RestoreNPMCache())
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem("install", workdir))

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.GenerateCordovaBuildConfigStepListItem())

	cordovaArchiveEnvs := []envmanModels.EnvironmentItemModel{
		{platformInputKey: "$" + platformInputEnvKey},
		{targetInputKey: targetEmulator},
	}
	if scanner.relCordovaConfigDir != "" {
		cordovaArchiveEnvs = append(cordovaArchiveEnvs, envmanModels.EnvironmentItemModel{workDirInputKey: "$" + workDirInputEnvKey})
	}
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.CordovaArchiveStepListItem(cordovaArchiveEnvs...))
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.SaveNPMCache())
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList()...)

	config, err := configBuilder.Generate(ScannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	return models.BitriseConfigMap{
		configName: string(data),
	}, nil
}

// DefaultConfigs ...
func (*Scanner) DefaultConfigs() (models.BitriseConfigMap, error) {
	configBuilder := models.NewDefaultConfigBuilder()
	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultPrepareStepList(steps.PrepareListParams{
		SSHKeyActivation: models.SSHKeyActivationConditional,
	})...)

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.CertificateAndProfileInstallerStepListItem())

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.RestoreNPMCache())

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.NpmStepListItem("install", "$"+workDirInputEnvKey))

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.GenerateCordovaBuildConfigStepListItem())

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.CordovaArchiveStepListItem(
		envmanModels.EnvironmentItemModel{workDirInputKey: "$" + workDirInputEnvKey},
		envmanModels.EnvironmentItemModel{platformInputKey: "$" + platformInputEnvKey},
		envmanModels.EnvironmentItemModel{targetInputKey: targetEmulator}))

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.SaveNPMCache())

	configBuilder.AppendStepListItemsTo(models.PrimaryWorkflowID, steps.DefaultDeployStepList()...)

	config, err := configBuilder.Generate(ScannerName)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return models.BitriseConfigMap{}, err
	}

	return models.BitriseConfigMap{
		defaultConfigName: string(data),
	}, nil
}
