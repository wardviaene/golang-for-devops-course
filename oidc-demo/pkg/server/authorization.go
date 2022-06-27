package server

import (
	"fmt"
	"net/http"
)

func (s server) authorization(w http.ResponseWriter, r *http.Request) {
	var (
		clientID     string
		redirectURI  string
		scope        string
		responseType string
		state        string
	)
	if clientID = r.URL.Query().Get("client_id"); clientID == "" {
		returnError(w, fmt.Errorf("client_id not supplied"))
		return
	}
	if redirectURI = r.URL.Query().Get("redirect_uri"); redirectURI == "" {
		returnError(w, fmt.Errorf("redirect_uri not supplied"))
		return
	}
	if scope = r.URL.Query().Get("scope"); scope == "" {
		returnError(w, fmt.Errorf("scope not supplied"))
		return
	}
	if responseType = r.URL.Query().Get("response_type"); responseType == "" {
		returnError(w, fmt.Errorf("response_type not supplied"))
		return
	}
	if state = r.URL.Query().Get("state"); state == "" {
		returnError(w, fmt.Errorf("state not supplied"))
		return
	}

}
