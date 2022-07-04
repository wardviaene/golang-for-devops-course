package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLoginGet(t *testing.T) {
	s := newServer(privkeyPem, testConfig)

	// 1. authorization flow
	endpoint := fmt.Sprintf("/authorization?client_id=%s&client_secret=%s&redirect_uri=%s&scope=openid&response_type=code&state=randomstring",
		s.Config.Apps["app1"].ClientID,
		s.Config.Apps["app1"].ClientSecret,
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

	// 2. login flow
	req = httptest.NewRequest(http.MethodGet, res.Header.Get("location"), nil)
	w = httptest.NewRecorder()
	s.login(w, req)
	loginRes := w.Result()
	defer loginRes.Body.Close()
	body, err := io.ReadAll(loginRes.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}
	if loginRes.StatusCode != http.StatusOK {
		t.Fatalf("HTTP StatusCode: %d (expected %d)", res.StatusCode, http.StatusAccepted)
	}

	if !strings.Contains(strings.ToLower(string(body)), "<html") {
		t.Fatalf("No HTML returned from login budy")
	}
}

func TestLoginPost(t *testing.T) {
	loginField := "login"
	passwordField := "password"

	loginValue := "edward"
	passwordValue := "password"

	s := newServer(privkeyPem, testConfig)

	// 1. authorization flow
	endpoint := fmt.Sprintf("/authorization?client_id=%s&client_secret=%s&redirect_uri=%s&scope=openid&response_type=code&state=randomstring",
		s.Config.Apps["app1"].ClientID,
		s.Config.Apps["app1"].ClientSecret,
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

	// 2. login get flow
	req = httptest.NewRequest(http.MethodGet, res.Header.Get("location"), nil)
	w = httptest.NewRecorder()
	s.login(w, req)
	loginRes := w.Result()
	defer loginRes.Body.Close()
	_, err = io.ReadAll(loginRes.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}
	if loginRes.StatusCode != http.StatusOK {
		t.Fatalf("HTTP StatusCode: %d (expected %d)", res.StatusCode, http.StatusAccepted)
	}

	// 3. Login post flow (we're also adding any values that were passed to the login page - in case a sessionID or state was added)
	loginUrl, err := url.Parse(res.Header.Get("location"))
	if err != nil {
		t.Fatalf("Couldn't parse loginUrl: %s", err)
	}
	form := url.Values{}
	form.Add(loginField, loginValue)
	form.Add(passwordField, passwordValue)
	for key, values := range loginUrl.Query() {
		for _, value := range values {
			form.Add(key, value)
		}
	}

	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	s.login(w, req)
	postLoginRes := w.Result()
	defer postLoginRes.Body.Close()
	_, err = io.ReadAll(postLoginRes.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}
	if postLoginRes.StatusCode != http.StatusFound {
		t.Fatalf("HTTP StatusCode: %d (expected %d).", res.StatusCode, http.StatusFound)
	}

	if postLoginRes.Header.Get("location") == "" {
		t.Fatalf("Location header not set")
	}

	postLoginUrl, err := url.Parse(postLoginRes.Header.Get("location"))
	if err != nil {
		t.Fatalf("Couldn't parse loginUrl: %s", err)
	}

	if postLoginUrl.Query().Get("code") == "" {
		t.Fatalf("No code received in login redirect")
	}

	fmt.Printf("Got location after login: %s\n", postLoginRes.Header.Get("location"))

}
