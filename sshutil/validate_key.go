package sshutil

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/retry"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// ValidatePrivateKey checks if can connect to a repository with a given private key
func ValidatePrivateKey(privateKey []byte, username string, url string) (bool, error) {
	SSHAuth, err := ssh.NewPublicKeys(username, privateKey, "")
	if err != nil {
		return false, err
	}
	var b bytes.Buffer
	if _, err = git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth:              SSHAuth,
		URL:               url,
		Progress:          &b,
		NoCheckout:        true,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}); err != nil {
		return false, err
	}
	log.Debugf(b.String())
	return true, nil
}

// SSHKeyPair is an asssymetric keypair used for SSH authentication
type SSHKeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// SSHRepo conmtains information to connect to an SSH git repository
type SSHRepo struct {
	Keys     SSHKeyPair
	Username string
	URL      string
}

// ValidateSSHAddedManually checks that a generated public key is added to the git service provider
func ValidateSSHAddedManually(repo SSHRepo) error {
	log.Warnf("Copy this SSH public key to your clipboard and add it to your Github repository or account!")
	fmt.Println(string(repo.Keys.PublicKey))

	return retry.Times(3).Try(func(attempt uint) error {
		log.Printf("Hit enter if you have finished with the setup")
		if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
			err = fmt.Errorf("failed to read line from input, error: %s", err)
			return nil
		}

		if valid, err := ValidatePrivateKey(repo.Keys.PrivateKey, repo.Username, repo.URL); !valid {
			log.Errorf("Could not connect to repository with private key, error: %s", err)
			return err
		}
		return nil
	})
}
