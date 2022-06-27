package server

import (
	"embed"
	"io/ioutil"
	"net/http"
)

//go:embed templates/*
var templateFs embed.FS

func (s server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Write([]byte("doing post"))
	} else {
		templateFile, err := templateFs.Open("templates/login.html")
		if err != nil {
			returnError(w, err)
			return
		}
		loginTemplate, err := ioutil.ReadAll(templateFile)
		if err != nil {
			returnError(w, err)
			return
		}
		w.Write(loginTemplate)
	}
}
