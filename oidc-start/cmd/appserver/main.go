package main

import (
	"fmt"
	"net/http"
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
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {

}
