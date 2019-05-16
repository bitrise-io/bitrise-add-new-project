package phases

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterWebhook(t *testing.T) {
	// happy path
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		if _, err := res.Write([]byte(`{"message":"ok"}`)); err != nil {
			t.Fatalf("write fake response: %s", err)
		}
	}))
	defer func() { testServer.Close() }()

	baseURL = testServer.URL
	appSlug := "dummy app slug"
	apiToken := "dummy token"

	if err := registerWebhook(appSlug, apiToken); err != nil {
		t.Fatalf("err should be nil instead of %s", err)
	}

	// fatal error
	testServer500 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(500)
	}))
	defer func() { testServer500.Close() }()

	baseURL = testServer500.URL

	err = registerWebhook(appSlug, apiToken)

	if err == nil {
		t.Fatalf("err should object instead of nil")
	}

	if webhookAttemptCount != 0 {
		t.Fatalf("multiple attemps detected")
	}
}
