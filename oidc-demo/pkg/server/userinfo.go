package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/users"
)

func (s *server) userinfo(w http.ResponseWriter, r *http.Request) {
	authorizationHeader := r.Header.Get("Authorization")

	if authorizationHeader == "" {
		returnError(w, fmt.Errorf("Authorization header empty"))
		return
	}

	authorizationHeader = strings.Replace(authorizationHeader, "Bearer ", "", -1)

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(authorizationHeader, claims, func(token *jwt.Token) (interface{}, error) {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("Parse private key error: %s", err)
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		returnError(w, fmt.Errorf("parse token error: %s", err))
		return
	}

	found := false
	for _, aud := range claims.Audience {
		if aud == s.Config.Url+"/userinfo" {
			found = true
		}
	}
	if !found {
		returnError(w, fmt.Errorf("token has incorrect audience: %s", strings.Join(claims.Audience, ", ")))
		return
	}
	if claims.Subject == "" {
		returnError(w, fmt.Errorf("subject is empty"))
		return
	}

	for _, user := range users.GetAllUsers() {
		if user.Sub == claims.Subject {
			out, err := json.Marshal(user)
			if err != nil {
				returnError(w, fmt.Errorf("json marshal error: %s", err))
				return
			}
			w.Write(out)
			return
		}
	}

	returnError(w, fmt.Errorf("user not found"))
	return
}
