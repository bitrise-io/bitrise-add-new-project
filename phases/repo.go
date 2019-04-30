package phases

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
)

type repoDetails struct {
	URL      string
	Provider string
	Owner    string
	Slug     string
	RepoType string
}

type urlParts struct {
	host  string
	owner string
	slug  string
}

func parseURL(cloneURL string) urlParts {
	parts := strings.SplitAfter(cloneURL, "https://")
	if len(parts) > 1 {
		// e.g. cloneURL=https://github.com/bitrise-io/go-utils.git
		parts = strings.Split(parts[1], "/")
		return urlParts{
			host:  parts[0],
			owner: parts[1],
			slug:  parts[2],
		}
	} else {
		// e.g. cloneURL=git@github.com:bitrise-io/go-utils.git
		afterAt := strings.SplitAfter(parts[0], "git@")[1]
		parts = strings.Split(afterAt, ":")
		host := parts[0]

		afterHost := strings.SplitAfter(afterAt, ":")[1]
		parts = strings.Split(afterHost, "/")
		return urlParts{
			host:  host,
			owner: parts[0],
			slug:  strings.TrimSuffix(parts[1], ".git"),
		}
	}
}

func buildURL(parts urlParts, protocol string) (cloneURL string) {
	switch protocol {
	case "https":
		cloneURL = fmt.Sprintf("https://%s/%s/%s.git", parts.host, parts.owner, parts.slug)
	case "ssh":
		cloneURL = fmt.Sprintf("git@%s:%s/%s.git", parts.host, parts.owner, parts.slug)
	}
	return
}

func getProvider(cloneURL string) (string) {
	if strings.Contains(cloneURL, "github.com") {
		return "github"
	} else if strings.Contains(cloneURL, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(cloneURL, "bitbucket.org") {
		return "bitbucket"
	}
	return "other"
}

// Repo returns repository details extracted from the working
// directory. If the Project visibility was set to public, the
// https clone url will be used.
func Repo(isPublic bool) (repoDetails, error) {
	log.Infof("SCANNING WORKDIR FOR GIT REPO")
	log.Infof("=============================")

	cmd := command.New("git", "remote", "get-url", "origin")
	log.Donef("$ %s", cmd.PrintableCommandArgs())
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return repoDetails{}, fmt.Errorf("get repo origin url: %s", out)
		} else {
			return repoDetails{}, fmt.Errorf("get repo origin url: %s", err)
		}
	}

	provider := getProvider(out)
	repoType := "git"

	parts := parseURL(out)
	var url string
	if isPublic {
		url = buildURL(parts, "https")
	} else {
		url = buildURL(parts, "ssh")
	}

	log.Donef("REPOSITORY SCANNED. DETAILS:")
	log.Donef("- url: %s", url)
	log.Donef("- provider: %s", provider)
	log.Donef("- owner: %s", parts.owner)
	log.Donef("- slug: %s", parts.slug)
	log.Donef("- repo type: %s", repoType)

	return repoDetails{
		url,
		provider,
		parts.owner,
		parts.slug,
		repoType,
	}, nil
}
