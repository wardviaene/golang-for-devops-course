package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-start/pkg/oidc"
)

func TestUserInfo(t *testing.T) {
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
		t.Fatalf("HTTP StatusCode: %d (expected %d)", res.StatusCode, http.StatusOK)
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
		t.Fatalf("HTTP StatusCode: %d (expected %d)", res.StatusCode, http.StatusFound)
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

	// 4. exchange code into token
	form = url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", s.Config.Apps["app1"].ClientID)
	form.Add("client_secret", s.Config.Apps["app1"].ClientSecret)
	form.Add("redirect_uri", s.Config.Apps["app1"].RedirectURIs[0])
	form.Add("code", postLoginUrl.Query().Get("code"))

	req = httptest.NewRequest(http.MethodPost, "/token", bytes.NewBufferString(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	s.token(w, req)
	tokenRes := w.Result()
	defer tokenRes.Body.Close()
	body, err := io.ReadAll(tokenRes.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}
	if tokenRes.StatusCode != http.StatusOK {
		t.Fatalf("HTTP StatusCode: %d (expected %d). Body: %s", tokenRes.StatusCode, http.StatusOK, body)
	}

	var tokenResponse oidc.Token

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		t.Errorf("Token Unmarshal error: %s", err)
	}

	if tokenResponse.IDToken == "" {
		fmt.Printf("IDToken is empty")
	}

	claims := jwt.StandardClaims{}
	_, err = jwt.ParseWithClaims(tokenResponse.IDToken, &claims, func(token *jwt.Token) (interface{}, error) {
		privateKeyParsed, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
		if err != nil {
			return nil, err
		}
		return &privateKeyParsed.PublicKey, nil
	})

	if err != nil {
		t.Fatalf("invalid token error: %s", err)
	}

	// 5. Use token for userinfo endpoint
	req = httptest.NewRequest(http.MethodPost, "/userinfo", bytes.NewBufferString(form.Encode()))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))
	w = httptest.NewRecorder()
	s.userinfo(w, req)
	userinfoRes := w.Result()
	defer userinfoRes.Body.Close()
	body, err = io.ReadAll(userinfoRes.Body)
	if err != nil {
		t.Errorf("Readall error: %s", err)
	}
	if userinfoRes.StatusCode != http.StatusOK {
		t.Fatalf("HTTP StatusCode: %d (expected %d). Body: %s", userinfoRes.StatusCode, http.StatusOK, body)
	}

	fmt.Printf("Got userinfo JSON: %s\n", body)

}
