package phases

import (
	"fmt"
	"strings"
)

type githubHandler struct {}

func (gh githubHandler) provider() string {
	return "github"
}

func (gh githubHandler) repoType() string {
	return "git"
}

func (gh githubHandler) parseURL(cloneURL string) urlParts {
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

func (gh githubHandler) buildURL(parts urlParts, protocol string) (cloneURL string) {
	switch protocol {
	case "https":
		cloneURL = fmt.Sprintf("https://%s/%s/%s.git", parts.host, parts.owner, parts.slug)
	case "ssh":
		cloneURL = fmt.Sprintf("git@%s:%s/%s.git", parts.host, parts.owner, parts.slug)
	}
	return
}