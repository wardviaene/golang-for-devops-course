package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
)

func main() {
	var (
		err error
	)
	privateKey, err := ioutil.ReadFile("mykey.pem")
	if err != nil {
		log.Fatalf("Failed to load mykey.pem, err: %v", err)
	}
	publicKey, err := ioutil.ReadFile("server.pub")
	if err != nil {
		log.Fatalf("Failed to load server.pub, err: %v", err)
	}

	privateKeyParsed, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to parse mykey.pem, err: %v", err)
	}
	publicKeyParsed, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		log.Fatalf("Failed to parse server.pub, err: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "username",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKeyParsed),
		},
		HostKeyCallback: ssh.FixedHostKey(publicKeyParsed),
	}
	client, err := ssh.Dial("tcp", "localhost:2022", config)
	if err != nil {
		log.Fatal("Failed to dial: ", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("NewSession error: ", err)
	}

	defer session.Close()

	out, err := session.Output("whoami")
	if err != nil {
		log.Fatal("Session output error: ", err)
	}

	fmt.Printf("Output is: %s\n", out)
}
