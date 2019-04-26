package phases

import (
	"fmt"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
)

type urlParts struct{
	host string
	owner string
	slug string
}

func getProviderHandler(cloneURL string) providerHandler {
	return githubHandler{}
}

// Repo ...
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

	handler := getProviderHandler(out)
	parts := handler.parseURL(out)


	var url string
	if isPublic {
		url = handler.buildURL(parts, "https")
	} else {
		url = handler.buildURL(parts, "ssh")
	}

	fmt.Printf("REPOSITORY SCANNED. DETAILS: url=%s provider=%s owner=%s slug=%s repoType=%s", url, handler.provider(), parts.owner, parts.slug, handler.repoType())
	fmt.Println()

	return url, handler.provider(), parts.owner, parts.slug, handler.repoType(), nil
}
