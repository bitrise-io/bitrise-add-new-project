package phases

import (
	"fmt"
	"net/url"
	"os"
	"strings"

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

func logRepoDetailsResult(repoURL RepoDetails) {
	log.Debugf("REPOSITORY SCANNED. DETAILS:")
	log.Debugf("- url: %s", repoURL.URL)
	log.Debugf("- provider: %s", repoURL.Provider)
	log.Debugf("- owner: %s", repoURL.Owner)
	log.Debugf("- slug: %s", repoURL.Slug)
	log.Debugf("- username: %s", repoURL.SSHUsername)
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
		Provider:    getProvider(URL.Hostname()),
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

func getProvider(hostName string) string {
	hostParts := strings.Split(hostName, ".")
	if len(hostParts) < 2 {
		return "other"
	}

	if hostParts[len(hostParts)-1] == "com" && hostParts[len(hostParts)-2] == "github" {
		return "github"
	} else if hostParts[len(hostParts)-1] == "com" && hostParts[len(hostParts)-2] == "gitlab" {
		return "gitlab"
	} else if hostParts[len(hostParts)-1] == "org" && hostParts[len(hostParts)-2] == "bitbucket" {
		return "bitbucket"
	}
	return "other"
}

func validateRepositoryAvailablePublic(url string) error {
	if _, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		Auth:              nil,
		URL:               url,
		Progress:          os.Stdout,
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
	log.Infof("SCANNING WORKDIR FOR GIT REPO")

	// Open local git repository
	repo, err := git.PlainOpen(searchDir)
	if err != nil {
		return RepoDetails{}, fmt.Errorf("failed to open git repository (%s), error: %s", searchDir, err)
	}

	log.Donef("Found git repository: %s", searchDir)

	// Get remote URL
	origin, err := repo.Remote("origin")
	if err != nil {
		return RepoDetails{}, fmt.Errorf("no remote 'origin' found in repository (%s), error: %s", searchDir, err)
	}

	if origin == nil || len(origin.Config().URLs) == 0 {
		return RepoDetails{}, fmt.Errorf("no URLs found for remote 'origin' in repository (%s)", searchDir)
	}
	remoteURL := origin.Config().URLs[0]

	log.Donef("Remote URL: %s", remoteURL)

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
			log.Donef("Repository (%s) is not public, error: %s", URL.String(), err)

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
		log.Infof("Checking if repository %s is public.", alternatePublicURL.String())

		if err := validateRepositoryAvailablePublic(alternatePublicURL.String()); err != nil {
			log.Infof("Alternate public URL is not available, error: %s", err)
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
			logRepoDetailsResult(*repoDetails)
			return *repoDetails, nil
		case SSHWithPublicAlternate:
			log.Donef("Using alternate public URL: %s", alternatePublicRepoDetails.URL)
			logRepoDetailsResult(*repoDetails)
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
		logRepoDetailsResult(*repoDetails)
		return *repoDetails, nil
	case SSHWithPublicAlternate:
		prompt := promptui.Select{
			Label: "Select repository URL:",
			Items: []string{alternatePublicRepoDetails.URL, repoDetails.URL},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return RepoDetails{}, fmt.Errorf("scan user input: %s", err)
		}
		
		if result == repoDetails.URL {
			logRepoDetailsResult(*repoDetails)
			return *repoDetails, nil
		}
		logRepoDetailsResult(*alternatePublicRepoDetails)
		return *alternatePublicRepoDetails, nil
	case HTTPSAuth:
		logRepoDetailsResult(*alternatePublicRepoDetails)
		return *alternateSSHRepoDetails, nil
	case SSH:
		logRepoDetailsResult(*repoDetails)
		return *repoDetails, nil
	default:
		return RepoDetails{}, fmt.Errorf("invalid state")
	}
}
