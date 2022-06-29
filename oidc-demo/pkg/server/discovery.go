package server

import (
	"encoding/json"
	"net/http"
)

func (s *server) discovery(w http.ResponseWriter, r *http.Request) {
	discovery := Discovery{
		Issuer:                            s.Config.Url,
		AuthorizationEndpoint:             s.Config.Url + "/authorization",
		TokenEndpoint:                     s.Config.Url + "/token",
		UserinfoEndpoint:                  s.Config.Url + "/userinfo",
		JwksURI:                           s.Config.Url + "/jwks.json",
		ScopesSupported:                   []string{"openid"},
		ResponseTypesSupported:            []string{"code"},
		TokenEndpointAuthMethodsSupported: []string{"none"},
	}
	out, err := json.Marshal(discovery)
	if err != nil {
		returnError(w, err)
		return
	}
	w.Write(out)
}
