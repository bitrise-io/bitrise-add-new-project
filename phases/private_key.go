package phases

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/pkg/errors"
)

type option struct {
	title        string
	valueOptions []string
	action       func(string) *option
}

func (o *option) run() {
	answer := ask(o.title, o.valueOptions...)
	if o.action != nil {
		if nextOption := o.action(answer); nextOption != nil {
			nextOption.run()
		}
	}
}

func ask(title string, options ...string) string {
	if len(options) == 1 {
		return options[0]
	}

	fmt.Print(strings.TrimSuffix(title, ":") + ":")

	if len(options) == 0 {
		fmt.Print(" ")
		for {
			input, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				log.Errorf("Error: failed to read input value")
				continue
			}
			fmt.Println()
			return strings.TrimSpace(input)
		}
	}

	fmt.Println()
	for i, option := range options {
		log.Printf("(%d) %s", i+1, option)
	}

	for {
		fmt.Print("Option number: ")
		answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Errorf("Error: failed to read input value")
			continue
		}
		optionNo, err := strconv.Atoi(strings.TrimSpace(answer))
		if err != nil {
			log.Errorf("Error: failed to parse option number, pick a number from 1-%d", len(options))
			continue
		}
		if optionNo-1 < 0 || optionNo-1 >= len(options) {
			log.Errorf("Error: invalid option number, pick a number 1-%d", len(options))
			continue
		}
		fmt.Println()
		return options[optionNo-1]
	}
}

func generateSSHKey() (string, string, error) {
	tempDir, err := pathutil.NormalizedOSTempDirPath("_key_")
	if err != nil {
		return "", "", err
	}

	keyFilePath := filepath.Join(tempDir, "key")

	if out, err := command.New("ssh-keygen", "-q", "-t", "rsa", "-b", "4096", "-C", "builds@bitrise.io", "-P", "", "-f", keyFilePath).RunAndReturnTrimmedCombinedOutput(); err != nil {
		return "", "", errors.Wrap(err, out)
	}

	return keyFilePath + ".pub", keyFilePath, nil
}

// PrivateKey ...
func PrivateKey() (string, string, bool, error) {
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
						return nil
					},
				}
			}
			return nil
		},
	}).run()

	return publicKeyPath, privateKeyPath, register, err
}
