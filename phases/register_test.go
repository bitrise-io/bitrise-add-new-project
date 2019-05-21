package phases

/*

func TestRegisterWebhook(t *testing.T) {
	// happy path
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		if _, err := res.Write([]byte(`{"message":"ok"}`)); err != nil {
			t.Fatalf("could not write fake response: %s", err)
		}
	}))
	defer func() { testServer.Close() }()

	baseURL = testServer.URL
	appSlug := "dummy app slug"
	apiToken := "dummy token"

	if err := registerWebhook(appSlug, apiToken); err != nil {
		t.Fatalf("expected: nil, got: error (%s)", err)
	}

	// fatal error
	testServer500 := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(500)
	}))
	defer func() { testServer500.Close() }()

	baseURL = testServer500.URL

	err := registerWebhook(appSlug, apiToken)

	if err == nil {
		t.Fatalf("expected: error, got: nil")
	}
}

*/
