package bitriseio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/log"
)

func TestAppsService_RegisterFinish(t *testing.T) {
	log.SetEnableDebugLog(true)

	// happy path
	{
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(200)
			r := RegisterFinishResponse{
				BuildTriggerToken: "trigger token",
			}
			b, err := json.Marshal(r)
			if err != nil {
				t.Errorf("could not marshal response: %s", err)
				return
			}
			if _, err := res.Write(b); err != nil {
				t.Errorf("could not write response: %s", err)
				return
			}
		}))
		defer func() { testServer.Close() }()

		baseURL = testServer.URL

		client, err := NewClient("api_token")
		if err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
		params := RegisterFinishParams{
			Config:      "default-ios-config",
			Envs:        nil,
			Mode:        "manual",
			ProjectType: "ios",
			StackID:     "osx-xcode-edge",
		}
		if resp, err := client.Apps.RegisterFinish("app_slug", params); err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		} else if resp == nil {
			t.Errorf("want: resp, got: %s", err)
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
		params := RegisterFinishParams{
			Config:      "default-ios-config",
			Envs:        nil,
			Mode:        "manual",
			ProjectType: "ios",
			StackID:     "osx-xcode-edge",
		}
		if resp, err := client.Apps.RegisterFinish("app_slug", params); err == nil {
			t.Errorf("want: error, got: %s", err)
			return
		} else if resp != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
	}
}
