package phases

import (
	"fmt"
	"strings"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

type urlParts struct{
	host string
	owner string
	slug string
}

func parseURL(cloneURL string) urlParts {
	parts := strings.SplitAfter(cloneURL, "https://")
	if len(parts) > 1 {
		// e.g. cloneURL=https://github.com/bitrise-io/go-utils.git
		parts = strings.Split(parts[1], "/")
		return urlParts{
			host: parts[0],
			owner: parts[1],
			slug: parts[2],
		}
	} else {
		// e.g. cloneURL=git@github.com:bitrise-io/go-utils.git
		afterAt := strings.SplitAfter(parts[0], "git@")[1]
		parts = strings.Split(afterAt, ":")
		host := parts[0]
		
		afterHost := strings.SplitAfter(afterAt, ":")[1]
		parts = strings.Split(afterHost, "/")
		return urlParts{
			host: host,
			owner: parts[0],
			slug: strings.TrimSuffix(parts[1], ".git"),
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

func getProvider(cloneURL string) string {
	if strings.Contains(cloneURL, "github.com") {
		return "github"
	} else if strings.Contains(cloneURL, "gitlab.com") {
		return "gitlab"
	} else if strings.Contains(cloneURL, "bitbucket.org") {
		return "bitbucket"
	}
	return ""
}

// Repo returns repository details extracted from the working
// directory. If the Project visibility was set to public, the
// https clone url will be used.
func Repo(isPublic bool) (string, string, string, string, string, error) {
	cmd := command.New("git", "remote", "get-url", "origin")
	out, err := cmd.RunAndReturnTrimmedCombinedOutput()
	if err != nil {
		if errorutil.IsExitStatusError(err) {
			return "", "", "", "", "", fmt.Errorf("get repo origin url: %s", out)
		} else {
			return "", "", "", "", "", fmt.Errorf("get repo origin url: %s", err)
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

	fmt.Printf("REPOSITORY SCANNED. DETAILS: url=%s provider=%s owner=%s slug=%s repoType=%s", url, provider, parts.owner, parts.slug, repoType)
	fmt.Println()

	return url, provider, parts.owner, parts.slug, repoType, nil
}
