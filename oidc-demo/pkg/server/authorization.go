package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

func (s *server) authorization(w http.ResponseWriter, r *http.Request) {

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

	// find appConfig
	var appConfig AppConfig
	for _, config := range s.Config.Apps {
		if config.ClientID == clientID {
			appConfig = config
		}
	}
	if appConfig.ClientID == "" {
		returnError(w, fmt.Errorf("clientID not recognized"))
		return
	}

	buf := make([]byte, 128)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		returnError(w, fmt.Errorf("crypto/rand is unavailable: Read() failed with %#v", err))
		return
	}

	sessionID := base64.URLEncoding.EncodeToString(buf)

	s.LoginRequests[sessionID] = LoginRequest{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		Scope:        scope,
		ResponseType: responseType,
		State:        state,
		AppConfig:    appConfig,
	}

	w.Header().Add("Location", "/login?sessionID="+sessionID)
	w.WriteHeader(http.StatusFound)
}
