package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-start/pkg/oidc"
)

func TestGetTokenFromCode(t *testing.T) {
	// generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		t.Fatalf("Couldn't generate rsa key")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// generate jwt and return
		if r.URL.Path == "/token" {
			claims := jwt.MapClaims{
				"iss": "http://example.com",
				"sub": "1-2-3-5",
				"aud": "1-2-3-4",
				"exp": time.Now().Add(1 * time.Hour).Unix(),
				"iat": time.Now().Unix(),
			}
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "0-0-0-1"

			idTokenString, err := token.SignedString(privateKey)
			if err != nil {
				t.Fatalf("SignedString: %s", err)
			}
			// generate Access Token
			claims = jwt.MapClaims{
				"iss": "http://example.com",
				"sub": "1-2-3-4",
				"aud": []string{
					"http://example.com/userinfo",
				},
				"exp": time.Now().Add(1 * time.Hour).Unix(),
				"iat": time.Now().Unix(),
			}
			token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "0-0-0-1"

			accessTokenString, err := token.SignedString(privateKey)
			if err != nil {
				t.Fatalf("SignedString: %s", err)
			}

			refreshToken, err := oidc.GetRandomString(64)
			if err != nil {
				t.Fatalf("random string error: %s", err)
			}
			responseToken := oidc.Token{
				IDToken:      idTokenString,
				AccessToken:  accessTokenString,
				ExpiresIn:    60,
				TokenType:    "bearer",
				RefreshToken: refreshToken,
			}

			out, err := json.Marshal(responseToken)
			if err != nil {
				t.Fatalf("json marshal error: %s", err)
				return
			}
			w.Write(out)
		} else if r.URL.Path == "/jwks.json" {
			jwks := oidc.Jwks{
				Keys: []oidc.JwksKey{
					{
						N:   base64.StdEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
						E:   "AQAB",
						Alg: "RS256",
						Use: "sig",
						Kid: "0-0-0-1",
						Kty: "RSA",
					},
				},
			}
			out, err := json.Marshal(jwks)
			if err != nil {
				t.Fatalf("jwks marshall error: %s", err)
				return
			}
			w.Write(out)
		}
	}))
	defer ts.Close()

	_, claims, err := getTokenFromCode(ts.URL+"/token", ts.URL+"/jwks.json", "http://localhost:8081", "1-2-3-4", "secret", "mycode")
	if err != nil {
		t.Fatalf("getTokenFromCode error: %s", err)
	}
	if claims == nil {
		t.Fatalf("claims is nil")
	}

	fmt.Printf("Exchanged code into token. Subject: %s\n", claims.Subject)

}
