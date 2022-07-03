package main

import "github.com/golang-jwt/jwt/v4"

// gets token from tokenUrl validating token with jwksUrl and returning token & claims
func getTokenFromCode(tokenUrl, jwksUrl, redirectUri, clientID, clientSecret, code string) (*jwt.Token, *jwt.StandardClaims, error) {
	return nil, nil, nil
}
