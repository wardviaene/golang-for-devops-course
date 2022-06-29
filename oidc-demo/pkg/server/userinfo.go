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
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		returnError(w, fmt.Errorf("authorization header not provided"))
		return
	}
	authHeader = strings.Replace(authHeader, "Bearer ", "", -1)
	claims := jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(authHeader, &claims, func(token *jwt.Token) (interface{}, error) {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(s.PrivateKey)
		if err != nil {
			return nil, err
		}
		return &privateKey.PublicKey, nil
	})
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(fmt.Sprintf("Token validation failed: %s", err)))
		return
	}

	for _, user := range users.GetAllUsers() {
		if user.Sub == claims.Subject {
			out, err := json.Marshal(user)
			if err != nil {
				returnError(w, fmt.Errorf("user marshal error: %s", err))
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
			return
		}
	}
	returnError(w, fmt.Errorf("user not found: %s", claims.Subject))
}
