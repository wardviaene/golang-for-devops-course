package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/server"
	"github.com/wardviaene/golang-for-devops-course/ssh-demo"
)

func main() {
	var (
		privateKey []byte
		err        error
	)
	// read config
	if _, err = os.Stat("config.yaml"); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Error: config.yaml doesn't exist\n")
		os.Exit(1)
	}
	config, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
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
	fmt.Printf("Server stopped: %s", server.Start(privateKey, server.ReadConfig(config)))
}
