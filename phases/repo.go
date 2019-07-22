package phases

import (
	"fmt"
	"net/url"

	// "os"
	"strings"

	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/log"
	git "gopkg.in/src-d/go-git.v4"
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
		Provider:    getProvider(URL.Hostname()),
	}, nil
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

	return *repoDetails, nil
}
