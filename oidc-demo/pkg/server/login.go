package server

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/users"
)

//go:embed templates/*
var templateFs embed.FS

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var (
			err          error
			ok           bool
			loginRequest LoginRequest
		)

		if err = r.ParseForm(); err != nil {
			returnError(w, err)
			return
		}

		if loginRequest, ok = s.LoginRequests[r.PostForm.Get("sessionID")]; !ok {
			returnError(w, fmt.Errorf("Invalid session ID: %s", r.PostForm.Get("sessionID")))
			fmt.Printf("loginRequests: %+v\n", s.LoginRequests)
			return
		}

		auth, user, err := users.Auth(r.PostForm.Get("login"), r.PostForm.Get("password"), "")
		if err != nil {
			returnError(w, err)
			return
		}
		if auth {
			code, err := oidc.GetRandomString(64)
			if err != nil {
				returnError(w, err)
				return
			}

			loginRequest.CodeIssuedAt = time.Now()
			loginRequest.User = user

			s.Codes[code] = loginRequest
			delete(s.LoginRequests, r.Form.Get("sessionid"))

			w.Header().Add("Location", loginRequest.RedirectURI+"?code="+code+"&state="+loginRequest.State)
			w.WriteHeader(http.StatusFound)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Authentication failed"))
		}

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
		loginTemplateStr := strings.Replace(string(loginTemplate), "$SESSIONID", r.URL.Query().Get("sessionID"), -1)
		w.Write([]byte(loginTemplateStr))
	}
}
