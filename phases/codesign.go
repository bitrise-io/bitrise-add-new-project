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

					codesigndoc, lerr := ioutil.TempFile("", "codesigndoc")
					if lerr != nil {
						err = lerr
						return nil
					}

					resp, lerr := http.Get("https://github.com/bitrise-io/codesigndoc/releases/download/2.3.0/codesigndoc-Darwin-x86_64")
					if lerr != nil {
						err = lerr
						return nil
					}

					if _, lerr = io.Copy(codesigndoc, resp.Body); lerr != nil {
						err = lerr
						return nil
					}

					if lerr := codesigndoc.Chmod(0700); lerr != nil {
						err = lerr
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

					debugSlice := codesignCmd
					debugSlice[3], debugSlice[5] = "*", "*"
					fmt.Println()
					log.Donef("%s", debugSlice)
					fmt.Println()

					cmd, lerr := command.NewFromSlice(codesignCmd)
					if lerr != nil {
						log.Errorf("%s", lerr)
						err = lerr
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
