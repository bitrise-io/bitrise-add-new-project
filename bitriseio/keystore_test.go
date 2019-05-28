package bitriseio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitrise-io/go-utils/log"
)

func TestUpload(t *testing.T) {
	log.SetEnableDebugLog(true)

	client, err := NewClient("UraOQmhccgEvSNiSvfdH_rRv7jn4XQVcOsID0_PqfWNdH88GTq-tT_VVwHUtve2nUVqt8eDRNZTiIB_vDpaVBw")
	if err != nil {
		t.Errorf("want: nil, got: %s", err)
		return
	}
	params := UploadKeystoreParams{
		Password:    "password",
		Alias:       "key0",
		KeyPassword: "password",
	}
	if err := client.Apps.UploadKeystore("2381434346e000d9", "/Users/godrei/Desktop/add-new-app-keystore.jks", params); err != nil {
		t.Errorf("want: nil, got: %s", err)
		return
	} else {
		t.Errorf("want: nil, got: %s", err)
		return
	}
}

func TestAppsService_UploadKeystore(t *testing.T) {
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
		params := UploadKeystoreParams{
			Password:    "",
			Alias:       "alias",
			KeyPassword: "",
		}
		if err := client.Apps.UploadKeystore("slug", "keystore.jks", params); err != nil {
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
		params := UploadKeystoreParams{
			Password:    "test",
			Alias:       "alias",
			KeyPassword: "test",
		}
		if err := client.Apps.UploadKeystore("slug", "keystore.jks", params); err == nil {
			t.Errorf("want: nil, got: %s", err)
			return
		}
	}
}
