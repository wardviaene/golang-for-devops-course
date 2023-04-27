package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthorization(t *testing.T) {
	s := newServer(privkeyPem, testConfig) // testConfig is defined in http_test.go and defines a static config

	endpoint := fmt.Sprintf("/authorization?client_id=%s&redirect_uri=%s&scope=openid&response_type=code&state=randomstring",
		s.Config.Apps["app1"].ClientID, // app1 is defined in testConfig
		s.Config.Apps["app1"].RedirectURIs[0],
	)
	req := httptest.NewRequest(http.MethodGet, endpoint, nil)
	w := httptest.NewRecorder()
	s.authorization(w, req)
	res := w.Result()
	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}

	if res.Header.Get("location") == "" {
		t.Fatalf("Location header not set")
	}

	if res.StatusCode != http.StatusFound {
		t.Fatalf("HTTP StatusCode: %d (expected %d)", res.StatusCode, http.StatusFound)
	}

	fmt.Printf("Got location: %s\n", res.Header.Get("location"))

}
