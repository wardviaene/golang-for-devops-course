package server

import (
	"fmt"
	"net/http"
)

type server struct {
	LoginRequests map[string]LoginRequest
	Codes         map[string]LoginRequest
}

func Start() error {
	s := &server{
		LoginRequests: make(map[string]LoginRequest),
		Codes:         make(map[string]LoginRequest),
	}
	http.HandleFunc("/authorization", s.authorization)
	http.HandleFunc("/token", s.token)
	http.HandleFunc("/login", s.login)
	http.HandleFunc("/.well-known/openid-configuration", s.discovery)

	return http.ListenAndServe(":8080", nil)
}

func returnError(w http.ResponseWriter, err error) {
	w.WriteHeader(400)
	w.Write([]byte(err.Error()))
	fmt.Printf("Error: %s\n", err)
}
