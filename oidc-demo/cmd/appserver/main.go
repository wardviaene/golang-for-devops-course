package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

const redirectUri = "http://localhost:8081/callback"

type app struct {
	states map[string]bool
}

func main() {

	a := app{
		states: make(map[string]bool),
	}

	http.HandleFunc("/", a.index)
	http.HandleFunc("/callback", a.callback)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("ListenAndServe error: %s\n", err)
	}
}

func (a *app) index(w http.ResponseWriter, r *http.Request) {
	oidcEndpoint := os.Getenv("OIDC_ENDPOINT")
	discovery, err := oidc.ParseDiscovery(oidcEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		returnError(w, fmt.Errorf("ParseDiscovery error: %s", err))
		return
	}

	state, err := oidc.GetRandomString(64)
	if err != nil {
		returnError(w, fmt.Errorf("GetRandomString error: %s", err))
		return
	}

	a.states[state] = true

	authorizationURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=openid&response_type=code&state=%s", discovery.AuthorizationEndpoint, os.Getenv("CLIENT_ID"), redirectUri, state)
	w.Write([]byte(`<html>
		<body>
		<a href="` + authorizationURL + `"><button style="width: 100px;">Login</button></a>
		</body>
	</html>`))
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	oidcEndpoint := os.Getenv("OIDC_ENDPOINT")
	discovery, err := oidc.ParseDiscovery(oidcEndpoint + "/.well-known/openid-configuration")
	if err != nil {
		returnError(w, fmt.Errorf("ParseDiscovery error: %s", err))
		return
	}

	if _, ok := a.states[r.URL.Query().Get("state")]; !ok {
		returnError(w, fmt.Errorf("state mismatch error"))
		return
	}

	delete(a.states, r.URL.Query().Get("state"))

	tokens, _, err := getTokenFromCode(discovery.TokenEndpoint, discovery.JwksURI, redirectUri, os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), r.URL.Query().Get("code"))
	if err != nil {
		returnError(w, fmt.Errorf("getTokenFromCode error: %s", err))
		return
	}

	req, err := http.NewRequest("GET", discovery.UserinfoEndpoint, nil)
	if err != nil {
		returnError(w, fmt.Errorf("newRequest error: %s", err))
		return
	}
	req.Header.Add("Authorization", "Bearer "+tokens[1].Raw)

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		returnError(w, fmt.Errorf("do request error: %s", err))
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		returnError(w, fmt.Errorf("ReadAll error: %s", err))
		return
	}

	fmt.Printf("IDToken: %s\n", tokens[0].Raw)

	w.Write([]byte(fmt.Sprintf("Token received. Userinfo: %s", body)))
}

func returnError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
	fmt.Printf("Error: %s\n", err)
}
