package bitriseio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/log"
)

func TestAppsService_TriggerBuild(t *testing.T) {
	log.SetEnableDebugLog(true)

	// happy path
	{
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(200)
		}))
		defer func() { testServer.Close() }()

		baseURL = testServer.URL

		client, err := NewClient("api_token")
		if err != nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
		params := TriggerBuildParams{
			BuildParams: BuildParams{
				Branch:     "master",
				WorkflowID: "primary",
			},
		}
		if err := client.Apps.TriggerBuild("app_slug", params); err != nil {
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
		params := TriggerBuildParams{
			BuildParams: BuildParams{
				Branch:     "master",
				WorkflowID: "primary",
			},
		}
		if err := client.Apps.TriggerBuild("app_slug", params); err == nil {
			t.Errorf("want: error, got: %s", err)
			return
		}
	}
}
