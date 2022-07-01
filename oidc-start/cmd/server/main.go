package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/wardviaene/golang-for-devops-course/oidc-start/pkg/server"
	"github.com/wardviaene/golang-for-devops-course/ssh-demo"
)

const configFile = ""

func main() {
	var (
		privateKey []byte
		err        error
	)
	// read config
	if _, err = os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Error: %s doesn't exist\n", configFile)
		os.Exit(1)
	}
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to load %s, err: %v", configFile, err)
	}
	// read encryption key
	if _, err = os.Stat("enckey.pem"); errors.Is(err, os.ErrNotExist) {
		if privateKey, _, err = ssh.GenerateKeys(); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		if err = os.WriteFile("enckey.pem", privateKey, 0600); err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
	} else {
		privateKey, err = ioutil.ReadFile("enckey.pem")
		if err != nil {
			log.Fatalf("Failed to load authorized_keys, err: %v", err)
		}

	}
	fmt.Printf("Server stopped: %s", server.Start(&http.Server{Addr: ":8080"}, privateKey, server.ReadConfig(config)))
}
