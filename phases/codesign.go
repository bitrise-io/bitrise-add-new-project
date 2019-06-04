package phases

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
)

// CodesignResult ...
type CodesignResult struct {
	KeystorePath, Password, Alias, KeyPassword string
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
func AutoCodesign(projectType, orgSlug, apiToken string) (CodesignResult, error) {
	log.Donef("Project type: %s", projectType)
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

					codesigndoc, err := ioutil.TempFile("", "codesigndoc")
					if err != nil {
						return nil
					}

					resp, err := http.Get("https://github.com/bitrise-io/codesigndoc/releases/download/latest/codesigndoc-Darwin-x86_64")
					if err != nil {
						return nil
					}

					if _, err = io.Copy(codesigndoc, resp.Body); err != nil {
						return nil
					}

					if err := codesigndoc.Chmod(0700); err != nil {
						return nil
					}

					codesignCmd := []string{
						codesigndoc.Name(),
						"scan",
						"--auth-token", apiToken,
						"--app-slug", orgSlug,
						"--write-files", "disable",
						"xcode",
					}
					cmd, err := command.NewFromSlice(codesignCmd)
					if err != nil {
						return nil
					}
					cmd.SetStderr(os.Stderr).SetStdin(os.Stdin).SetStdout(os.Stdout)

					if err = cmd.Run(); err != nil {
						err = fmt.Errorf("failed to run codesigndoc: %s, error: %s", cmd.PrintableCommandArgs(), err)
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
