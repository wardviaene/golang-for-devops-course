package oidc

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v4"
)

// helper function to get Token from token endpoint
func GetTokenFromCode(tokenUrl, jwksUrl, redirectUri, clientID, clientSecret, code string) (*jwt.Token, *jwt.StandardClaims, error) {
	return nil, nil, nil
}

// helper function to get *rsa.PublicKey from a Jwks URL
func getPublicKeyFromJwks(jwksUrl, kid string) (*rsa.PublicKey, error) {
	return nil, nil
}
