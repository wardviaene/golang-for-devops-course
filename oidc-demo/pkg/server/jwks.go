package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

func (s *server) jwks(w http.ResponseWriter, r *http.Request) {

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
	if err != nil {
		returnError(w, fmt.Errorf("private key parsing error: %s", err))
		return
	}

	publicKey := privateKey.PublicKey

	jwks := oidc.Jwks{
		Keys: []oidc.JwksKey{
			{
				Kid: "0-0-0-1",
				Alg: "RS256",
				Kty: "RSA",
				Use: "sig",
				N:   base64.StdEncoding.EncodeToString(publicKey.N.Bytes()),
				E:   "AQAB",
			},
		},
	}
	out, err := json.Marshal(jwks)
	if err != nil {
		returnError(w, fmt.Errorf("jwks marshal error: %s", err))
		return
	}
	w.Write(out)
}
