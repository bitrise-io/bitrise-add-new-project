package phases

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/bitrise-add-new-project/sshutil"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/pkg/errors"
)

func readPrivateKey(keyFilePath string) ([]byte, error) {
	privateKey, err := fileutil.ReadStringFromFile(keyFilePath)
	if err != nil {
		return nil, fmt.Errorf("SSH private key read failed: %s", err)
	}
	privateKey = strings.TrimSuffix(privateKey, "\n")
	privateKey = strings.Replace(privateKey, "OPENSSH", "RSA", -1)
	return []byte(privateKey), nil
}

func generateSSHKey() (sshutil.SSHKeyPair, error) {
	tempDir, err := pathutil.NormalizedOSTempDirPath("_key_")
	if err != nil {
		return sshutil.SSHKeyPair{}, err
	}

	keyFilePath := filepath.Join(tempDir, "key")

	cmd := command.New("ssh-keygen", "-q", "-t", "rsa", "-b", "2048", "-C", "builds@bitrise.io", "-P", "", "-f", keyFilePath, "-m", "PEM")
	if out, err := cmd.RunAndReturnTrimmedCombinedOutput(); err != nil {
		return sshutil.SSHKeyPair{}, errors.Wrap(fmt.Errorf("failed to run command: %s, error: %s", cmd.PrintableCommandArgs(), err), out)
	}

	privateKey, err := readPrivateKey(keyFilePath)
	if err != nil {
		return sshutil.SSHKeyPair{}, err
	}

	publicKey, err := fileutil.ReadStringFromFile(keyFilePath + ".pub")
	if err != nil {
		return sshutil.SSHKeyPair{}, fmt.Errorf("SSH public key read failed: %s", err)
	}

	return sshutil.SSHKeyPair{
		PublicKey:  []byte(publicKey),
		PrivateKey: privateKey,
	}, nil
}

// PrivateKey ...
func PrivateKey(repoURL RepoDetails) (sshutil.SSHKeyPair, bool, error) {
	log.Infof("Setup repository access")
	fmt.Println()

	var (
		err      error
		register bool
		SSHKeys  sshutil.SSHKeyPair
	)

	const (
		methodTitle  = "Specify how Bitrise will be able to access the source code"
		methodAuto   = "Automatic (Git provider must be connected at: https://app.bitrise.io/me/profile or will fall back to manual registration.)"
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
				SSHKeys, err = generateSSHKey()
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
							register = false

							err = sshutil.ValidateSSHAddedManually(sshutil.SSHRepo{
								Keys:     SSHKeys,
								URL:      repoURL.URL,
								Username: repoURL.SSHUsername,
							})
							return nil
						}
						return nil
					},
				}
			case methodManual:
				register = false
				SSHKeys.PublicKey = nil

				err = retry.Times(3).Try(func(attempt uint) error {
					var privateKeyPath string
					(&option{
						title: privateKeyPathTitle,
						action: func(answer string) *option {
							privateKeyPath, err = pathutil.AbsPath(answer)
							if err != nil {
								log.Errorf("could not expand path (%s) to full path: %s", answer, err)
							}
							return nil
						},
					}).run()

					SSHKeys.PrivateKey, err = readPrivateKey(privateKeyPath)
					if err != nil {
						return err
					}

					var valid bool
					if valid, err = sshutil.ValidatePrivateKey(SSHKeys.PrivateKey, repoURL.SSHUsername, repoURL.URL); !valid {
						log.Errorf("Could not connect to repository with private key, error: %s", err)
						return err
					}
					return nil
				})
			}
			return nil
		},
	}).run()

	return SSHKeys, register, err
}
