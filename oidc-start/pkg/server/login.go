package server

import (
	"embed"
	"net/http"
)

//go:embed templates/*
var templateFs embed.FS

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	// to access the login template:
	// templateFile, err := templateFs.Open("templates/login.html")
}
