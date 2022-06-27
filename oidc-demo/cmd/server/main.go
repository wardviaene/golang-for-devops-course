package main

import (
	"fmt"

	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/server"
)

func main() {
	fmt.Printf("Server stopped: %s", server.Start())
}
