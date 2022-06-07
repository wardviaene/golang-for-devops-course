package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/wardviaene/golang-for-devops-course/ssh"
)

func main() {
	var (
		err error
	)
	authorizedKeysBytes, err := ioutil.ReadFile("mykey.pub")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	if err = ssh.StartServer(authorizedKeysBytes); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

}
