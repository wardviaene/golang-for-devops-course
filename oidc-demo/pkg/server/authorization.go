package server

import (
	"fmt"
	"net/http"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
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
		returnError(w, fmt.Errorf("client_id is empty"))
		return
	}
	if redirectURI = r.URL.Query().Get("redirect_uri"); redirectURI == "" {
		returnError(w, fmt.Errorf("redirectURI is empty"))
		return
	}
	if scope = r.URL.Query().Get("scope"); scope == "" {
		returnError(w, fmt.Errorf("scope is empty"))
		return
	}
	if responseType = r.URL.Query().Get("response_type"); responseType != "code" {
		returnError(w, fmt.Errorf("response_type is empty"))
		return
	}
	if state = r.URL.Query().Get("state"); state == "" {
		returnError(w, fmt.Errorf("state is empty"))
		return
	}
	appConfig := AppConfig{}
	for _, app := range s.Config.Apps {
		if app.ClientID == clientID {
			appConfig = app
		}
	}
	if appConfig.ClientID == "" {
		returnError(w, fmt.Errorf("client_id not found"))
		return
	}

	found := false
	for _, redirectURIConfig := range appConfig.RedirectURIs {
		if redirectURIConfig == redirectURI {
			found = true
		}
	}
	if !found {
		returnError(w, fmt.Errorf("redirect_uri not whitelisted"))
		return
	}

	sessionID, err := oidc.GetRandomString(128)
	if err != nil {
		returnError(w, fmt.Errorf("GetRandomString error: %s", err))
		return
	}

	s.LoginRequest[sessionID] = LoginRequest{
		ClientID:     clientID,
		RedirectURI:  redirectURI,
		Scope:        scope,
		ResponseType: responseType,
		State:        state,
		AppConfig:    appConfig,
	}

	w.Header().Add("location", fmt.Sprintf("/login?sessionID=%s", sessionID))
	w.WriteHeader(http.StatusFound)

}
