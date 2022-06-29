package oidc

import (
	"encoding/json"
	"io"
	"net/http"
)

type Discovery struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

func ParseDiscovery(url string) (Discovery, error) {
	var discovery Discovery
	res, err := http.Get(url)
	if err != nil {
		return discovery, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err = json.Unmarshal(body, &discovery); err != nil {
		return discovery, err
	}
	return discovery, nil
}
