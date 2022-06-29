package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func (s *server) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		returnError(w, err)
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
		"sub": loginData.User.Login,
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
		"sub": loginData.User.Login,
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

	buf := make([]byte, 64)

	_, err = io.ReadFull(rand.Reader, buf)
	if err != nil {
		returnError(w, fmt.Errorf("crypto/rand is unavailable: Read() failed with %#v", err))
		return
	}

	refreshToken := base64.URLEncoding.EncodeToString(buf)

	responseToken := Token{
		IDToken:      idTokenString,
		AccessToken:  accessTokenString,
		ExpiresIn:    60,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
	}
	out, err := json.Marshal(responseToken)
	if err != nil {
		returnError(w, fmt.Errorf("json marshal error: %s", err))
		return
	}
	w.Write(out)
}
