package phases

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"
)

func generateSSHKey() (string, string, error) {
	tempDir, err := pathutil.NormalizedOSTempDirPath("_key_")
	if err != nil {
		return "", "", err
	}

	keyFilePath := filepath.Join(tempDir, "key")

	cmd := command.New("ssh-keygen", "-q", "-t", "rsa", "-b", "2048", "-C", "builds@bitrise.io", "-P", "", "-f", keyFilePath, "-m", "PEM")
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return "", "", errors.Wrap(fmt.Errorf("failed to run command: %s, error: %s", cmd.PrintableCommandArgs(), err), out)
	}

	return keyFilePath + ".pub", keyFilePath, nil
}

func validatePrivateKey(path string, url string) (bool, error) {
	return false, nil
}

// PrivateKey ...
func PrivateKey(repoURL string) (string, string, bool, error) {
	log.Infof("Setup repository access")
	fmt.Println()

	var (
		err                           error
		register                      bool
		publicKeyPath, privateKeyPath string
	)

	const (
		methodTitle  = "Specify how Bitrise will be able to access the source code"
		methodAuto   = "Automatic"
		methodManual = "Add own SSH"
	)
	(&option{
		title:        methodTitle,
		valueOptions: []string{methodAuto, methodManual},
		action: func(answer string) *option {
			const (
				privateKeyPathTitle   = "Enter the path of your RSA SSH private key file (you can also drag & drop the file here)"
				additionalAccessTitle = "Do you need to use an additional private repository?"
				additionalAccessNo    = "No, auto-add SSH key"
				additionalAccessYes   = "I need to"
			)
			switch answer {
			case methodAuto:
				register = true
				publicKeyPath, privateKeyPath, err = generateSSHKey()
				if err != nil {
					return nil
				}
				return &option{
					title:        additionalAccessTitle,
					valueOptions: []string{additionalAccessNo, additionalAccessYes},
					action: func(answer string) *option {
						switch answer {
						case additionalAccessNo:
							return nil
						case additionalAccessYes:
							log.Warnf("Copy this SSH public key to your clipboard and add it to your Github repository or account!")
							content, readErr := ioutil.ReadFile(publicKeyPath)
							if readErr != nil {
								err = readErr
								return nil
							}
							fmt.Println(string(content))
							return &option{
								title: "Hit enter if you have finished with the setup",
								action: func(_ string) *option {
									return nil
								},
							}
						}
						return nil
					},
				}
			case methodManual:
				register = false
				publicKeyPath = ""
				return &option{
					title: privateKeyPathTitle,
					action: func(answer string) *option {
						privateKeyPath = answer

						if ok, err := validatePrivateKey(privateKeyPath, repoURL); !ok {
							log.Errorf("Private key invalid: %s", err)
							return nil
						}
						
						log.Printf("Private key valid!")
						return nil
					},
				}
			}
			return nil
		},
	}).run()

	return publicKeyPath, privateKeyPath, register, err
}
