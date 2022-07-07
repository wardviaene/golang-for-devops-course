package server

import (
	"time"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/users"
)

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

type LoginRequest struct {
	ClientID     string
	RedirectURI  string
	Scope        string
	ResponseType string
	State        string
	CodeIssuedAt time.Time
	User         users.User
	AppConfig    AppConfig
}
