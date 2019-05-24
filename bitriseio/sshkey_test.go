package bitriseio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/log"
)

func TestAppsService_RegisterSSHKey(t *testing.T) {
	log.SetEnableDebugLog(true)

	// happy path
	{
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(200)
			if _, err := res.Write([]byte(`{"message":"ok"}`)); err != nil {
				t.Fatalf("could not write fake response: %s", err)
			}
		}))
		defer func() { testServer.Close() }()

		baseURL = testServer.URL

		client, err := NewClient("api_token")
		if err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
		params := RegisterSSHKeyParams{
			AuthSSHPrivateKey:                `ssh private key`,
			AuthSSHPublicKey:                 `ssh public key`,
			IsRegisterKeyIntoProviderService: true,
		}
		if err := client.Apps.RegisterSSHKey("app_slug", params); err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
	}

	// fatal error
	{
		testServer500 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(500)
		}))
		defer func() { testServer500.Close() }()

		baseURL = testServer500.URL

		client, err := NewClient("api_token")
		if err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
		params := RegisterSSHKeyParams{
			AuthSSHPrivateKey:                `ssh private key`,
			AuthSSHPublicKey:                 `ssh public key`,
			IsRegisterKeyIntoProviderService: true,
		}
		if err := client.Apps.RegisterSSHKey("app_slug", params); err == nil {
			t.Errorf("want: error, got: %s", err)
			return
		}
	}
}
