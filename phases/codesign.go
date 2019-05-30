package phases

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
)

// CodesignResult ...
type CodesignResult struct {
	KeystorePath, Password, Alias, KeyPassword string
	ProfilePaths, CertificatePaths             []string
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
func AutoCodesign(projectType string) (CodesignResult, error) {
	log.Donef("Found %s based project", projectType)
	fmt.Println()

	var result CodesignResult
	var err error
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
				if projectType == platformIOS || projectType == platformBoth {

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

				if projectType == platformAndroid || projectType == platformBoth {
					(&option{
						title: "Enter key store path",
						action: func(answer string) *option {
							result.KeystorePath = answer
							return nil
						}}).run()

					(&option{
						title:  "Enter key store password",
						secret: true,
						action: func(answer string) *option {
							result.Password = answer
							return nil
						}}).run()

					(&option{
						title: "Enter key alias",
						action: func(answer string) *option {
							result.Alias = answer
							return nil
						}}).run()

					(&option{
						title:  "Enter key password",
						secret: true,
						action: func(answer string) *option {
							result.KeyPassword = answer
							return nil
						}}).run()
				}

				return nil
			}
			return nil
		}}).run()

	return result, err
}
