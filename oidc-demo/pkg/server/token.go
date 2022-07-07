package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

func (s *server) token(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		returnError(w, fmt.Errorf("Not a POST request"))
		return
	}
	if err := r.ParseForm(); err != nil {
		returnError(w, fmt.Errorf("ParseForm error: %s", err))
		return
	}
	if r.PostForm.Get("grant_type") != "authorization_code" {
		returnError(w, fmt.Errorf("invalid grant type: %s", r.PostForm.Get("grant_type")))
		return
	}
	loginRequest, ok := s.Codes[r.PostForm.Get("code")]
	if !ok {
		returnError(w, fmt.Errorf("invalid code"))
		return
	}
	if time.Now().After(loginRequest.CodeIssuedAt.Add(10 * time.Minute)) {
		returnError(w, fmt.Errorf("code expired"))
		return
	}
	if loginRequest.ClientID != r.PostForm.Get("client_id") {
		returnError(w, fmt.Errorf("client_id mismatch"))
		return
	}
	if loginRequest.AppConfig.ClientSecret != r.PostForm.Get("client_secret") {
		returnError(w, fmt.Errorf("invalid client_secret"))
		return
	}
	if loginRequest.RedirectURI != r.PostForm.Get("redirect_uri") {
		returnError(w, fmt.Errorf("invalid redirect_uri"))
		return
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
	if err != nil {
		returnError(w, fmt.Errorf("private key parsing error: %s", err))
		return
	}
	claims := jwt.MapClaims{
		"iss": s.Config.Url,
		"sub": loginRequest.User.Sub,
		"aud": loginRequest.ClientID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0-0-0-1"

	signedIDToken, err := token.SignedString(privateKey)
	if err != nil {
		returnError(w, fmt.Errorf("signedString error: %s", err))
		return
	}

	// access token
	claims = jwt.MapClaims{
		"iss": s.Config.Url,
		"sub": loginRequest.User.Sub,
		"aud": []string{
			s.Config.Url + "/userinfo",
		},
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0-0-0-1"

	signedAccessToken, err := token.SignedString(privateKey)
	if err != nil {
		returnError(w, fmt.Errorf("signedString error: %s", err))
		return
	}

	tokenOutput := oidc.Token{
		AccessToken: signedAccessToken,
		IDToken:     signedIDToken,
		TokenType:   "bearer",
		ExpiresIn:   60,
	}

	delete(s.Codes, r.PostForm.Get("code"))

	out, err := json.Marshal(tokenOutput)
	if err != nil {
		returnError(w, fmt.Errorf("token marshal error: %s", err))
		return
	}
	w.Write(out)
}
