#!/bin/sh
# ./start-test-server.sh
# You can use this script to start the server
go get "github.com/golang-jwt/jwt/v4"
go run assignment1.go main.go ratelimit.go
