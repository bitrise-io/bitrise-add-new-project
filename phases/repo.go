package phases

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/goinp/goinp"
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

type urlParts struct {
	host        string
	owner       string
	slug        string
	scheme      RepoScheme
	SSHUsername string
	URL         *url.URL
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
	if len(pathParts) != 2 {
		return urlParts{}, fmt.Errorf("URL path does not contain exactly two parts")
	}

	return urlParts{
		scheme:      scheme,
		host:        parsed.Hostname(),
		owner:       pathParts[0],
		slug:        strings.TrimRight(pathParts[len(pathParts)-1], ".git"),
		SSHUsername: SSHUsername,
		URL:         parsed,
	}, nil
}

func setSchemeToHTTPS(URL *url.URL) *url.URL {
	URL.Scheme = "https"
	return URL
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
	parts, err := parseURL(remoteURL)
	if err != nil {
		return RepoDetails{}, err
	}
	provider := getProvider(parts.host)

	repoDetails := RepoDetails{
		Scheme:      parts.scheme,
		URL:         parts.URL.String(),
		Provider:    provider,
		Owner:       parts.owner,
		Slug:        parts.slug,
		SSHUsername: parts.SSHUsername,
	}

	// Validate https repositoy
	if parts.scheme == HTTPS {
		if err := validateRepositoryAvailablePublic(parts.URL.String()); err != nil {
			return RepoDetails{}, fmt.Errorf("could not check repository (%s) with git clone, error: %s", parts.URL.String(), err)
		}
	}

	// If ssh repository is provided, check the alternate availability with https scheme
	var alternatePublicRepoDetails *RepoDetails
	if parts.scheme == SSH {
		alternatePublicURL := setSchemeToHTTPS(parts.URL)

		if err := validateRepositoryAvailablePublic(alternatePublicURL.String()); err != nil {
			log.Debugf("Alternate public URL is not available, error: %s", err)
		} else {
			alternatePublicRepoDetails = &RepoDetails{
				Scheme:   HTTPS,
				URL:      alternatePublicURL.String(),
				Provider: provider,
				Owner:    parts.owner,
				Slug:     parts.slug,
			}
		}
	}

	// Public Bitrise app
	if isPublicApp {
		if parts.URL.Scheme == "https" {
			return repoDetails, nil
		}
		// scheme is SSH

		// Public app authenticated clone URL is not allowed
		if alternatePublicRepoDetails != nil {
			return RepoDetails{}, fmt.Errorf(("public Bitrise app must use a public git repository but have SSH clone URL"))
		}

		log.Donef("Using alternate public URL: %s", alternatePublicRepoDetails.URL)
		return *alternatePublicRepoDetails, nil
	}

	// Private Bitrise app
	if parts.URL.Scheme == "ssh" && alternatePublicRepoDetails != nil {
		result, err := goinp.AskOptions("Select repository URL:", alternatePublicRepoDetails.URL, false, []string{alternatePublicRepoDetails.URL, repoDetails.URL}...)
		if err != nil {
			return RepoDetails{}, err
		}

		if result == repoDetails.URL {
			return repoDetails, nil
		}
		return *alternatePublicRepoDetails, nil
	}

	return repoDetails, nil
}
