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
