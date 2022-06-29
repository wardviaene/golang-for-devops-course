package server

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

func (s *server) jwks(w http.ResponseWriter, r *http.Request) {
	var (
		parsedKey any
		ok        bool
		err       error
	)
	block, _ := pem.Decode(s.PrivateKey)
	if err != nil {
		returnError(w, err)
		return
	}

	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		returnError(w, err)
		return
	}

	var privateKey *rsa.PrivateKey
	if privateKey, ok = parsedKey.(*rsa.PrivateKey); !ok {
		returnError(w, fmt.Errorf("unable to parse public key"))
		return
	}
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
		returnError(w, err)
		return
	}
	w.Write(out)
}
