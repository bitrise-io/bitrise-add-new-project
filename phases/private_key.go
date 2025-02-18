package phases

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/bitrise-io/bitrise-add-new-project/sshutil"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/ssh"
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
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return sshutil.SSHKeyPair{}, err
	}

	var privateKeyPEM bytes.Buffer
	privateKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privateKeyPEM, privateKeyBlock); err != nil {
		return sshutil.SSHKeyPair{}, err
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return sshutil.SSHKeyPair{}, err
	}

	publicKeyString := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(publicKey)), "\n")
	publicKeyString = publicKeyString + " builds@bitrise.io\n"

	return sshutil.SSHKeyPair{
		PrivateKey: privateKeyPEM.Bytes(),
		PublicKey:  []byte(publicKeyString),
	}, nil
}

// PrivateKey ...
func PrivateKey(repoURL RepoDetails) (sshutil.SSHKeyPair, bool, error) {
	fmt.Println()
	log.Infof("SETUP REPOSITORY ACCESS")
	log.Printf("For automatic ssh key registration git provider must be connected at: https://app.bitrise.io/me/profile")

	var (
		err      error
		register bool
		SSHKeys  sshutil.SSHKeyPair
	)

	const (
		methodTitle  = "Specify how Bitrise will be able to access the source code"
		methodAuto   = "Automatic"
		methodManual = "Add own SSH"
	)

	prompt := promptui.Select{
		Label: methodTitle,
		Items: []string{methodAuto, methodManual},
		Templates: &promptui.SelectTemplates{
			Selected: "Repo access method: {{ . | green }}",
		},
	}

	_, method, err := prompt.Run()
	if err != nil {
		return SSHKeys, false, fmt.Errorf("scan user input: %s", err)
	}

	if method == methodAuto {
		if SSHKeys, err = generateSSHKey(); err != nil {
			return SSHKeys, false, err
		}

		const (
			additionalAccessTitle = "Do you need to use an additional private repository?"
			additionalAccessNo    = "No, auto-add SSH key"
			additionalAccessYes   = "I need to"
		)
		prompt := promptui.Select{
			Label: additionalAccessTitle,
			Items: []string{additionalAccessNo, additionalAccessYes},
			Templates: &promptui.SelectTemplates{
				Label:    fmt.Sprintf("%s {{.}} ", promptui.IconInitial),
				Selected: "Need to add additional repo: {{ . | green }}",
			},
		}

		_, additional, err := prompt.Run()
		if err != nil {
			return SSHKeys, false, fmt.Errorf("scan user input: %s", err)
		}

		if additional == additionalAccessNo {
			return SSHKeys, true, nil
		}

		log.Warnf("Copy this SSH public key to your clipboard and add it to any additional Git repository or account!")
		fmt.Println(string(SSHKeys.PublicKey))

		log.Printf("Hit enter if you have finished with the setup")
		if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
			return SSHKeys, false, fmt.Errorf("failed to read line from input, error: %s", err)
		}

		return SSHKeys, true, nil
	}

	const privateKeyPathTitle = "Enter the path of your RSA SSH private key file (you can also drag & drop the file here)"

	register = false
	SSHKeys.PublicKey = nil

	err = retry.Times(3).Try(func(attempt uint) error {
		prompt := promptui.Prompt{
			Label: privateKeyPathTitle,
			Templates: &promptui.PromptTemplates{
				Success: "Private key path: {{ . | green }}",
			},
		}

		privateKeyPath, err := prompt.Run()
		if err != nil {
			return fmt.Errorf("scan user input: %s", err)
		}

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

	return SSHKeys, register, err
}
