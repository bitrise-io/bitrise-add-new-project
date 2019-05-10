package phases

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterWebhook(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte(`{"message":"ok"}`))
	}))
	defer func() { testServer.Close() }()

	baseURL = testServer.URL
	appSlug := "dummy app slug"
	apiToken := "dummy token"

	err := registerWebhook(appSlug, apiToken)

	if err != nil {
		t.Fatalf("err should be nil instead of %s", err)
	}

	// testServer400 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
	// 	res.WriteHeader(400)
	// }))
	// defer func() { testServer400.Close() }()

	// baseURL = testServer400.URL

	// err = registerWebhook(appSlug, apiToken)

	// if err == nil {
	// 	t.Fatalf("err should object instead of nil")
	// }
}