package phases

import (
	"fmt"
	"net/url"
	"strings"

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

type urlParts struct {
	host        string
	owner       string
	slug        string
	scheme      RepoScheme
	SSHUsername string
}

func parseURL(cloneURL string) (urlParts, error) {
	cloneURL = strings.TrimSpace(cloneURL)
	const pathSeperator = "/"

	if strings.HasPrefix(cloneURL, "git@") { // e.g. git@github.com:bitrise-io/go-utils.git
		cloneURL = strings.Replace(cloneURL, ":", pathSeperator, 1)
		cloneURL = "ssh://" + cloneURL
	}

	// Supporting the formats:
	// ssh://git@github.com/bitrise-io/go-utils.git
	// https://github.com/bitrise-io/go-utils.git

	parsed, err := url.Parse(cloneURL)
	if err != nil {
		return urlParts{}, err
	}

	var scheme RepoScheme
	switch parsed.Scheme {
	case "https":
		scheme = HTTPS
	case "ssh":
		scheme = SSH
	default:
		scheme = Invalid
	}
	if scheme == Invalid {
		return urlParts{}, fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}

	if scheme == HTTPS && parsed.User.Username() != "" {
		return urlParts{}, fmt.Errorf("username or access token is included in https git repository, only public https repositories are supported")
	}
	var SSHUsername string
	if scheme == SSH {
		SSHUsername = parsed.User.Username()
	}

	escapedPath := strings.TrimPrefix(parsed.EscapedPath(), pathSeperator)
	pathParts := strings.Split(escapedPath, pathSeperator)
	log.Debugf("URL path parts: %s", pathParts)
	if len(pathParts) < 2 {
		return urlParts{}, fmt.Errorf("URL path does not contain at least two elements")
	}

	return urlParts{
		scheme:      scheme,
		host:        parsed.Hostname(),
		owner:       pathParts[0],
		slug:        strings.TrimRight(pathParts[len(pathParts)-1], ".git"),
		SSHUsername: SSHUsername,
	}, nil
}

func buildURL(parts urlParts, ssh bool) string {
	if ssh {
		return fmt.Sprintf("git@%s:%s/%s.git", parts.host, parts.owner, parts.slug)
	}
	return fmt.Sprintf("https://%s/%s/%s.git", parts.host, parts.owner, parts.slug)
}

func getProvider(hostName string) string {
	if strings.HasSuffix(hostName, "github.com") {
		return "github"
	} else if strings.HasSuffix(hostName, "gitlab.com") {
		return "gitlab"
	} else if strings.HasSuffix(hostName, "bitbucket.org") {
		return "bitbucket"
	}
	return "other"
}

// Repo returns repository details extracted from the working
// directory. If the Project visibility was set to public, the
// https clone url will be used.
func Repo(searchDir string, isPublic bool) (RepoDetails, error) {
	log.Infof("SCANNING WORKDIR FOR GIT REPO")
	log.Infof("=============================")

	repo, err := git.PlainOpen(searchDir)
	if err != nil {
		return RepoDetails{}, fmt.Errorf("failed to open git repository (%s), error: %s", searchDir, err)
	}
	origin, err := repo.Remote("origin")
	if err != nil {
		return RepoDetails{}, fmt.Errorf("No remote 'origin' found, error: %s", err)
	}

	var remoteURL string
	if origin != nil && len(origin.Config().URLs) != 0 {
		remoteURL = origin.Config().URLs[0]
	}

	parts, err := parseURL(remoteURL)
	if err != nil {
		return RepoDetails{}, err
	}
	provider := getProvider(parts.host)

	var url string
	if isPublic {
		url = buildURL(parts, false)
	} else {
		url = buildURL(parts, true)
	}

	log.Donef("REPOSITORY SCANNED. DETAILS:")
	log.Donef("- url: %s", url)
	log.Donef("- provider: %s", provider)
	log.Donef("- owner: %s", parts.owner)
	log.Donef("- slug: %s", parts.slug)

	return RepoDetails{
		Scheme:      parts.scheme,
		URL:         url,
		Provider:    provider,
		Owner:       parts.owner,
		Slug:        parts.slug,
		SSHUsername: parts.SSHUsername,
	}, nil
}
