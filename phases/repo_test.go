package phases

import (
	"reflect"
	"testing"
)

func Test_parseURL(t *testing.T) {
	tests := []struct {
		name     string
		cloneURL string
		want     urlParts
		wantErr  bool
	}{
		{
			name:     "ssh with git@ prefix",
			cloneURL: "git@github.com:bitrise-io/go-utils.git",
			want: urlParts{
				scheme:      SSH,
				host:        "github.com",
				owner:       "bitrise-io",
				slug:        "go-utils",
				SSHUsername: "git",
			},
		},
		{
			name:     "https",
			cloneURL: "https://github.com/bitrise-io/go-utils.git",
			want: urlParts{
				scheme: HTTPS,
				host:   "github.com",
				owner:  "bitrise-io",
				slug:   "go-utils",
			},
		},
		{
			name:     "https with username or access token",
			cloneURL: "https://token@github.com/bitrise-io/go-utils.git",
			want:     urlParts{},
			wantErr:  true,
		},
		{
			name:     "https with username and passwork",
			cloneURL: "https://username:pass@github.com/bitrise-io/go-utils.git",
			want:     urlParts{},
			wantErr:  true,
		},
		{
			name:     "ssh with ssh:// prefix",
			cloneURL: "ssh://git@github.com/bitrise-io/go-utils.git",
			want: urlParts{
				scheme:      SSH,
				host:        "github.com",
				owner:       "bitrise-io",
				slug:        "go-utils",
				SSHUsername: "git",
			},
		},
		{
			name:     "filepath",
			cloneURL: "../f",
			want:     urlParts{},
			wantErr:  true,
		},
		{
			name:     "filepath",
			cloneURL: "file://test",
			want:     urlParts{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURL(tt.cloneURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseURL() returned err: %s, wantErr: %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
