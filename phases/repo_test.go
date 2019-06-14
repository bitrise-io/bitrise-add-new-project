package phases

import (
	"net/url"
	"reflect"
	"testing"
)

func Test_getProvider(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name     string
		hostName string
		want     string
	}{
		{
			hostName: "github.com",
			want:     "github",
		},
		{
			hostName: "d.github.com",
			want:     "github",
		},
		{
			hostName: "bitbucket.org",
			want:     "bitbucket",
		},
		{
			hostName: "github.com.unknown.com",
			want:     "other",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProvider(tt.hostName); got != tt.want {
				t.Errorf("getProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseURL(t *testing.T) {
	const httpsURL = "https://github.com/bitrise-io/go-utils.git"
	const sshURL = "ssh://git@github.com/bitrise-io/go-utils.git"

	tests := []struct {
		name     string
		cloneURL string
		want     string
		wantErr  bool
	}{
		{
			name:     "ssh with git@ prefix",
			cloneURL: "git@github.com:bitrise-io/go-utils.git",
			want:     "ssh://git@github.com/bitrise-io/go-utils.git",
		},
		{
			name:     "https with username or access token",
			cloneURL: "https://token@github.com/bitrise-io/go-utils.git",
			want:     "https://github.com/bitrise-io/go-utils.git",
		},
		{
			name:     "https with username and password",
			cloneURL: "https://user:pass@github.com/bitrise-io/go-utils.git",
			want:     "https://github.com/bitrise-io/go-utils.git",
		},
		{
			name:     "https",
			cloneURL: httpsURL,
			want:     httpsURL,
		},
		{
			name:     "ssh scheme URL with username",
			cloneURL: sshURL,
			want:     sshURL,
		},
		{
			name:     "filepath",
			cloneURL: "../f",
			want:     "../f",
		},
		{
			name:     "file scheme",
			cloneURL: "file://test",
			want:     "file://test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURL(tt.cloneURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("parseURL() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func Test_splitURL(t *testing.T) {
	sshURL, err := url.Parse("ssh://git@github.com/bitrise-io/go-utils.git")
	if err != nil {
		t.Errorf("setup: failed to parse URL")
	}

	httpsURL, err := url.Parse("https://github.com/bitrise-io/go-utils.git")
	if err != nil {
		t.Errorf("setup: failed to parse URL")
	}

	httpsAuthURL, err := url.Parse("https://token@github.com/bitrise-io/go-utils.git")
	if err != nil {
		t.Errorf("setup: failed to parse URL")
	}

	tests := []struct {
		name    string
		URL     *url.URL
		want    urlParts
		wantErr bool
	}{
		{
			name: "SSH URL",
			URL:  sshURL,
			want: urlParts{
				scheme:      SSH,
				owner:       "bitrise-io",
				slug:        "go-utils",
				SSHUsername: "git",
			},
		},
		{
			name: "HTTPS URL",
			URL:  httpsURL,
			want: urlParts{
				scheme:      HTTPS,
				owner:       "bitrise-io",
				slug:        "go-utils",
				SSHUsername: "",
			},
		},
		{
			name: "HTTPS Auth URL",
			URL:  httpsAuthURL,
			want: urlParts{
				scheme:      HTTPS,
				owner:       "bitrise-io",
				slug:        "go-utils",
				SSHUsername: "token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitURL(tt.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
