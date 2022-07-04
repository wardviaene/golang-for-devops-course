package server

import (
	"net/http"
)

type server struct {
	PrivateKey []byte
	Config     Config
}

func newServer(privateKey []byte, config Config) *server {
	return &server{
		PrivateKey: privateKey,
		Config:     config,
	}
}

func Start(httpServer *http.Server, privateKey []byte, config Config) error {
	s := newServer(privateKey, config)

	http.HandleFunc("/authorization", s.authorization)
	http.HandleFunc("/token", s.token)
	http.HandleFunc("/login", s.login)
	http.HandleFunc("/jwks.json", s.jwks)
	http.HandleFunc("/.well-known/openid-configuration", s.discovery)
	http.HandleFunc("/userinfo", s.userinfo)

	return httpServer.ListenAndServe()
}
