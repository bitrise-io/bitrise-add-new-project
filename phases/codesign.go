package phases

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/go-yaml/yaml"
)

// CodesignResult ...
type CodesignResult struct {
	KeystorePath, KeystorePassword, Alias, AliasPassword string
	ProfilePaths, CertificatePaths                       []string
}

const (
	platformIOS        = "iOS"
	platformAndroid    = "Android"
	platformBoth       = "iOS and Android"
	codesignExportsDir = "codesigndoc_exports"
)

func getPlatform(projectType string) string {
	switch projectType {
	case "ios":
		return platformIOS
	case "android":
		return platformAndroid
	case "xamarin", "flutter", "cordova", "ionic", "react-native":
		return platformBoth
	case "web", "macos":
		// not supported
	}
	// other or if no project type specified
	return ""
}

// AutoCodesign ...
func AutoCodesign(bitriseYMLPath string) (CodesignResult, error) {
	bitriseYMLContent, err := ioutil.ReadFile(bitriseYMLPath)
	if err != nil {
		return CodesignResult{}, fmt.Errorf("failed to open "+bitriseYMLPath+", error: %s", err)
	}

	var bitriseYML models.BitriseDataModel
	if err := yaml.Unmarshal(bitriseYMLContent, &bitriseYML); err != nil {
		return CodesignResult{}, fmt.Errorf("failed to parse "+bitriseYMLPath+", error: %s", err)
	}

	platform := getPlatform(bitriseYML.ProjectType)
	if platform == "" {
		log.Warnf("No project type set or unknown platform found.")
		return CodesignResult{}, nil
	}

	log.Donef("Found %s based project", platform)
	fmt.Println()

	var result CodesignResult

	const (
		exportTitle = "Do you want to export and upload codesigning files?"
		exportYes   = "Yes"
		exportNo    = "No"
	)
	(&option{
		title:        exportTitle,
		valueOptions: []string{exportYes, exportNo},
		action: func(answer string) *option {
			if answer == exportYes {
				if platform == platformIOS || platform == platformBoth {

					log.Infof("Exporting iOS codesigning files")

					var resp *http.Response
					resp, err = http.Get("https://raw.githubusercontent.com/bitrise-tools/codesigndoc/master/_scripts/install_wrap-xcode.sh")
					if err != nil {
						return nil
					}

					if _, err = io.Copy(os.Stdin, resp.Body); err != nil {
						return nil
					}

					cmd := command.New("bash").SetStderr(os.Stderr).SetStdin(os.Stdin).SetStdout(os.Stdout)
					if err = cmd.Run(); err != nil {
						err = fmt.Errorf("failed to run command: %s, error: %s", cmd.PrintableCommandArgs(), err)
					}

					var files []os.FileInfo
					files, err = ioutil.ReadDir(codesignExportsDir)
					if err != nil {
						return nil
					}

					for _, file := range files {
						switch ext := strings.HasSuffix; {
						case ext(file.Name(), ".p12"):
							result.CertificatePaths = append(result.CertificatePaths, filepath.Join(codesignExportsDir, file.Name()))
						case ext(file.Name(), ".mobileprovision"):
							result.ProfilePaths = append(result.ProfilePaths, filepath.Join(codesignExportsDir, file.Name()))
						}
					}
				}

				if platform == platformAndroid || platform == platformBoth {
					(&option{
						title: "Enter keystore path",
						action: func(answer string) *option {
							result.KeystorePath = answer
							return nil
						}}).run()

					(&option{
						title:  "Enter keystore password",
						secret: true,
						action: func(answer string) *option {
							result.KeystorePassword = answer
							return nil
						}}).run()

					(&option{
						title: "Enter keystore alias",
						action: func(answer string) *option {
							result.Alias = answer
							return nil
						}}).run()

					(&option{
						title:  "Enter keystore alias password",
						secret: true,
						action: func(answer string) *option {
							result.AliasPassword = answer
							return nil
						}}).run()
				}

				return nil
			}
			return nil
		}}).run()

	return result, err
}
