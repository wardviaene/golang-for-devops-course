package server

import (
	"time"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/users"
)

type LoginRequest struct {
	ResponseType string
	RedirectURI  string
	ClientID     string
	Scope        string
	State        string
	CodeIssuedAt time.Time
	User         users.User
	AppConfig    AppConfig
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

type Jwks struct {
	Keys []JwksKey `json:"keys"`
}
type JwksKey struct {
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
}

type Config struct {
	Apps      map[string]AppConfig `yaml:"apps"`
	Url       string               `yaml:"url"`
	LoadError error
}
type AppConfig struct {
	ClientID     string   `yaml:"clientID"`
	ClientSecret string   `yaml:"clientSecret"`
	Issuer       string   `yaml:"issuer"`
	RedirectURIs []string `yaml:"redirectURIs"`
}
