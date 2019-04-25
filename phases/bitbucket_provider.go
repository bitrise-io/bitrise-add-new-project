package phases

import (
	"fmt"
	"strings"
)

type bitbucketHandler struct{}

func (bh bitbucketHandler) provider() string {
	return "bitbucket"
}

func (bh bitbucketHandler) parseURL(cloneURL string) urlParts {
	parts := strings.SplitAfter(cloneURL, "https://")
	if len(parts) > 1 {
		// e.g. cloneURL=https://bitrise_lszucs@bitbucket.org/bitrise_lszucs/test-repo.git
		afterAt := strings.SplitAfter(parts[1], "@")[1]
		parts := strings.Split(afterAt, "/")

		return urlParts{
			host: parts[0],
			owner: parts[1],
			slug: strings.TrimSuffix(parts[2], ".git"),
		}
	} else {
		// e.g. cloneURL=git@bitbucket.org:bitrise_lszucs/test-repo.git
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

func (bh bitbucketHandler) buildURL(parts urlParts, protocol string) (cloneURL string) {
	switch protocol {
	case "https":
		cloneURL = fmt.Sprintf("https://%s/%s/%s.git", parts.host, parts.owner, parts.slug)
	case "ssh":
		cloneURL = fmt.Sprintf("ssh://git@%s:%s/%s.git", parts.host, parts.owner, parts.slug)
	}
	return
}