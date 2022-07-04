package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

const redirectUri = "http://localhost:8081/callback"

type app struct {
}

func main() {

	a := app{}

	http.HandleFunc("/", a.index)
	http.HandleFunc("/callback", a.callback)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("ListenAndServe error: %s\n", err)
	}
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	discovery, err := oidc.ParseDiscovery(os.Getenv("DISCOVERY_URL"))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("ParseDiscovery error: %s", err)))
		return
	}
	state, err := oidc.GetRandomString(32)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("GetRandomString error: %s", err)))
		return
	}
	authUrl := fmt.Sprintf("%s?client_id=%s&scope=openid&redirect_uri=%s&response_type=code&state=%s", discovery.AuthorizationEndpoint, os.Getenv("CLIENT_ID"), redirectUri, state)
	w.Write(getLoginButton(authUrl))
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("code") == "" {
		w.Write([]byte("Code not supplied"))
		return
	}
	discovery, err := oidc.ParseDiscovery(os.Getenv("DISCOVERY_URL"))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("ParseDiscovery error: %s", err)))
		return
	}
	_, claims, err := getTokenFromCode(discovery.TokenEndpoint, discovery.JwksURI, redirectUri, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), r.URL.Query().Get("code"))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("GetTokenFromCode error: %s", err)))
		return
	}

	w.Write([]byte(fmt.Sprintf("Got token. Sub: %s", claims.Subject)))
}

func getLoginButton(authUrl string) []byte {
	return []byte(`<html>
	<body>
		<a href="` + authUrl + `"><button style="width: 100px;">Login</button></a>
	</body>
	</html>`)
}
