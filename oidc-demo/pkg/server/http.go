package server

import (
	"fmt"
	"net/http"
)

type server struct {
	LoginRequests map[string]LoginRequest
	Codes         map[string]LoginRequest
	PrivateKey    []byte
	Config        Config
}

func Start(privateKey []byte, config Config) error {
	s := &server{
		LoginRequests: make(map[string]LoginRequest),
		Codes:         make(map[string]LoginRequest),
		PrivateKey:    privateKey,
		Config:        config,
	}

	if len(config.Apps) == 0 {
		return fmt.Errorf("No apps loaded, check your config (%s)", config.LoadError)
	}

	http.HandleFunc("/authorization", s.authorization)
	http.HandleFunc("/token", s.token)
	http.HandleFunc("/login", s.login)
	http.HandleFunc("/jwks.json", s.jwks)
	http.HandleFunc("/.well-known/openid-configuration", s.discovery)

	return http.ListenAndServe(":8080", nil)
}

func returnError(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	w.Write([]byte(err.Error()))
	fmt.Printf("Error: %s\n", err)
}
