package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

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
	authUrl := fmt.Sprintf("%s?client_id=%s&scope=openid&redirect_uri=%s&response_type=code&state=%s", discovery.AuthorizationEndpoint, os.Getenv("CLIENT_ID"), "http://localhost:8081/callback", state)
	w.Write(getLoginButton(authUrl))
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("parse code"))
}

func getLoginButton(authUrl string) []byte {
	return []byte(`<html>
	<body>
		<a href="` + authUrl + `"><button style="width: 100px;">Login</button></a>
	</body>
	</html>`)
}