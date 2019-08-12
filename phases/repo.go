package phases

import (
	"bytes"
	"fmt"
	"net/url"

	// "os"
	"strings"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	"github.com/manifoldco/promptui"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// RepoScheme is the type of the git repository protocol
type RepoScheme int

const (
	// Invalid is an unsupported git repo scheme type
	Invalid RepoScheme = iota
	// HTTPS is the https git repo scheme type
	HTTPS
	// SSH is the ssh git repo scheme type
	SSH
)

// RepoDetails encapsulates data needed to perform
// repo registration related requests through the
// Bitrise API
type RepoDetails struct {
	URL         string
	Provider    string
	Owner       string
	Slug        string
	Scheme      RepoScheme
	SSHUsername string
}

const urlPathSeperator = "/"

func parseURL(cloneURL string) (*url.URL, error) {
	cloneURL = strings.TrimSpace(cloneURL)

	if strings.HasPrefix(cloneURL, "git@") { // e.g. git@github.com:bitrise-io/go-utils.git
		cloneURL = strings.Replace(cloneURL, ":", urlPathSeperator, 1)
		cloneURL = "ssh://" + cloneURL
	}

	// Supporting the formats:
	// ssh://git@github.com/bitrise-io/go-utils.git
	// https://github.com/bitrise-io/go-utils.git

	parsed, err := url.Parse(cloneURL)
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "https" {
		if parsed.User.Username() != "" {
			log.Debugf("username or access token is included in https git repository")
		}
		parsed.User = nil
	}

	return parsed, nil
}

func splitURL(URL *url.URL) (*RepoDetails, error) {
	var scheme RepoScheme
	switch URL.Scheme {
	case "https":
		scheme = HTTPS
	case "ssh":
		scheme = SSH
	default:
		return &RepoDetails{}, fmt.Errorf("unsupported URL scheme: %s", URL.Scheme)
	}

	escapedPath := strings.TrimPrefix(URL.EscapedPath(), urlPathSeperator)
	pathParts := strings.Split(escapedPath, urlPathSeperator)
	log.Debugf("URL path parts: %s", pathParts)
	if len(pathParts) < 2 {
		return &RepoDetails{}, fmt.Errorf("URL path does not contain at least two parts")
	}

	return &RepoDetails{
		URL:         URL.String(),
		Scheme:      scheme,
		Owner:       pathParts[0],
		Slug:        strings.TrimRight(pathParts[len(pathParts)-1], ".git"),
		SSHUsername: URL.User.Username(),
		Provider:    "custom",
	}, nil
}

func schemeToHTTPS(URL *url.URL) *url.URL {
	httpsURL := &url.URL{}
	*httpsURL = *URL
	httpsURL.Scheme = "https"
	httpsURL.User = nil
	return httpsURL
}

func schemeToSSH(URL *url.URL) *url.URL {
	sshURL := &url.URL{}
	*sshURL = *URL
	sshURL.Scheme = "ssh"
	sshURL.User = url.User("git")
	return sshURL
}

func validateRepositoryAvailablePublic(url string) error {
	if _, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth:              nil,
		URL:               url,
		Progress:          bytes.NewBuffer([]byte{}),
		NoCheckout:        true,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}); err != nil {
		return err
	}
	return nil
}

// Repo returns repository details extracted from the working
// directory. If the Project visibility was set to public, the
// https clone url will be used.
func Repo(searchDir string, isPublicApp bool) (RepoDetails, error) {
	fmt.Println()
	log.Infof("SCANNING GIT REPOSITORY")

	// Open local git repository
	repo, err := git.PlainOpen(searchDir)
	if err != nil {
		return RepoDetails{}, fmt.Errorf("failed to open git repository (%s), error: %s", searchDir, err)
	}

	log.Debugf("Found git repository: %s", searchDir)

	// Get remote URL
	origin, err := repo.Remote("origin")
	if err != nil {
		return RepoDetails{}, fmt.Errorf("no remote 'origin' found in repository (%s), error: %s", searchDir, err)
	}

	if origin == nil || len(origin.Config().URLs) == 0 {
		return RepoDetails{}, fmt.Errorf("no URLs found for remote 'origin' in repository (%s)", searchDir)
	}
	remoteURL := origin.Config().URLs[0]

	log.Printf("Remote URL: %s", colorstring.Green(remoteURL))

	// Parse remote URL
	URL, err := parseURL(remoteURL)
	if err != nil {
		return RepoDetails{}, err
	}

	repoDetails, err := splitURL(URL)
	if err != nil {
		return RepoDetails{}, err
	}

	// Validate https repositoy
	var alternateSSHRepoDetails *RepoDetails
	if repoDetails.Scheme == HTTPS {
		if err := validateRepositoryAvailablePublic(URL.String()); err != nil {
			log.Warnf("Repository (%s) is not public, error: %s", URL.String(), err)

			var err error
			if alternateSSHRepoDetails, err = splitURL(schemeToSSH(URL)); err != nil {
				return RepoDetails{}, err
			}
		}
	}

	// If ssh repository is provided, check the alternate availability with https scheme
	var alternatePublicRepoDetails *RepoDetails
	if repoDetails.Scheme == SSH {
		alternatePublicURL := schemeToHTTPS(URL)
		log.Debugf("Checking if repository %s is public.", alternatePublicURL.String())

		if err := validateRepositoryAvailablePublic(alternatePublicURL.String()); err != nil {
			log.Warnf("Alternate public URL is not available, error: %s", err)
		} else {
			var err error
			if alternatePublicRepoDetails, err = splitURL(alternatePublicURL); err != nil {
				return RepoDetails{}, err
			}
		}
	}

	type repoAuth int
	const (
		Invalid repoAuth = iota
		HTTPSPublic
		HTTPSAuth
		SSHWithPublicAlternate
		SSH
	)

	var auth repoAuth
	if repoDetails.Scheme == HTTPS {
		if alternateSSHRepoDetails != nil {
			auth = HTTPSAuth
		} else {
			auth = HTTPSPublic
		}
	} else {
		if alternatePublicRepoDetails != nil {
			auth = SSHWithPublicAlternate
		} else {
			auth = SSH
		}
	}

	// Public Bitrise app
	if isPublicApp {
		switch auth {
		case HTTPSPublic:
			return *repoDetails, nil
		case SSHWithPublicAlternate:
			log.Debugf("Using alternate public URL: %s", alternatePublicRepoDetails.URL)
			return *alternatePublicRepoDetails, nil
		case HTTPSAuth:
			fallthrough
		case SSH:
			// Public app authenticated clone URL is not allowed
			return RepoDetails{}, fmt.Errorf(("public Bitrise app must use a git repository without authentication"))
		default:
			return RepoDetails{}, fmt.Errorf("invalid state")
		}
	}

	// Private Bitrise app
	switch auth {
	case HTTPSPublic:
		return *repoDetails, nil
	case SSHWithPublicAlternate:
		prompt := promptui.Select{
			Label: "Select repository URL:",
			Items: []string{alternatePublicRepoDetails.URL, repoDetails.URL},
			Templates: &promptui.SelectTemplates{
				Selected: "Selected repository: {{ . | green }}",
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return RepoDetails{}, fmt.Errorf("scan user input: %s", err)
		}

		if result == repoDetails.URL {
			return *repoDetails, nil
		}
		return *alternatePublicRepoDetails, nil
	case HTTPSAuth:
		return *alternateSSHRepoDetails, nil
	case SSH:
		return *repoDetails, nil
	default:
		return RepoDetails{}, fmt.Errorf("invalid state")
	}
	return *repoDetails, nil
}
