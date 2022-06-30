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
		returnError(w, fmt.Errorf("Got %s request instead of POST", r.Method))
		return
	}

	if err := r.ParseForm(); err != nil {
		returnError(w, err)
		return
	}

	if r.PostForm.Get("grant_type") != "authorization_code" {
		returnError(w, fmt.Errorf("grant_type must be authorization_code"))
		return
	}

	loginData, ok := s.Codes[r.PostForm.Get("code")]
	if !ok {
		returnError(w, fmt.Errorf("Code not supplied"))
		return
	}
	if loginData.CodeIssuedAt.After(time.Now().Add(10 * time.Minute)) {
		returnError(w, fmt.Errorf("Code expired"))
		return
	}

	if loginData.AppConfig.ClientID != r.PostForm.Get("client_id") {
		returnError(w, fmt.Errorf("client_id mismatch"))
		return
	}
	if loginData.AppConfig.ClientSecret != r.PostForm.Get("client_secret") {
		returnError(w, fmt.Errorf("incorect client secret"))
		return
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
	if err != nil {
		returnError(w, fmt.Errorf("ParseRSAPrivateKeyFromPEM: %s", err))
		return
	}

	// generate id Token
	claims := jwt.MapClaims{
		"iss": loginData.AppConfig.Issuer,
		"sub": loginData.User.Sub,
		"aud": loginData.ClientID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0-0-0-1"

	idTokenString, err := token.SignedString(privateKey)
	if err != nil {
		returnError(w, fmt.Errorf("SignedString: %s", err))
		return
	}

	// generate Access Token
	claims = jwt.MapClaims{
		"iss": loginData.AppConfig.Issuer,
		"sub": loginData.User.Sub,
		"aud": loginData.ClientID,
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "0-0-0-1"

	accessTokenString, err := token.SignedString(privateKey)
	if err != nil {
		returnError(w, fmt.Errorf("SignedString: %s", err))
		return
	}

	refreshToken, err := oidc.GetRandomString(64)
	if err != nil {
		returnError(w, err)
		return
	}
	responseToken := oidc.Token{
		IDToken:      idTokenString,
		AccessToken:  accessTokenString,
		ExpiresIn:    60,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
	}

	// remove code
	delete(s.Codes, r.PostForm.Get("code"))

	out, err := json.Marshal(responseToken)
	if err != nil {
		returnError(w, fmt.Errorf("json marshal error: %s", err))
		return
	}
	w.Write(out)
}
