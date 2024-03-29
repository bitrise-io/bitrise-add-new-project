package macos

import (
	"github.com/bitrise-io/bitrise-init/models"
	"github.com/bitrise-io/bitrise-init/scanners/ios"
)

//------------------
// ScannerInterface
//------------------

// Scanner ...
type Scanner struct {
	detectResult ios.DetectResult

	configDescriptors []ios.ConfigDescriptor
}

// NewScanner ...
func NewScanner() *Scanner {
	return &Scanner{}
}

// Name ...
func (Scanner) Name() string {
	return string(ios.XcodeProjectTypeMacOS)
}

// DetectPlatform ...
func (scanner *Scanner) DetectPlatform(searchDir string) (bool, error) {
	result, err := ios.ParseProjects(ios.XcodeProjectTypeMacOS, searchDir, true, false)
	if err != nil {
		return false, err
	}

	scanner.detectResult = result
	detected := len(result.Projects) > 0
	return detected, err
}

// ExcludedScannerNames ...
func (Scanner) ExcludedScannerNames() []string {
	return []string{}
}

// Options ...
func (scanner *Scanner) Options() (models.OptionNode, models.Warnings, models.Icons, error) {
	options, configDescriptors, _, warnings, err := ios.GenerateOptions(ios.XcodeProjectTypeMacOS, scanner.detectResult)
	if err != nil {
		return models.OptionNode{}, warnings, nil, err
	}

	scanner.configDescriptors = configDescriptors

	return options, warnings, nil, nil
}

// DefaultOptions ...
func (Scanner) DefaultOptions() models.OptionNode {
	return ios.GenerateDefaultOptions(ios.XcodeProjectTypeMacOS)
}

// Configs ...
func (scanner *Scanner) Configs(isPrivateRepository bool) (models.BitriseConfigMap, error) {
	return ios.GenerateConfig(ios.XcodeProjectTypeMacOS, scanner.configDescriptors, isPrivateRepository)
}

// DefaultConfigs ...
func (Scanner) DefaultConfigs() (models.BitriseConfigMap, error) {
	return ios.GenerateDefaultConfig(ios.XcodeProjectTypeMacOS)
}
