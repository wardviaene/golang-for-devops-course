package server

type Config struct {
	Apps      map[string]AppConfig
	Url       string
	LoadError error
}
type AppConfig struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	RedirectURIs []string
}
