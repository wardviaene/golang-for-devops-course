package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"github.com/golang-jwt/jwt/v4"
	"github.com/wardviaene/golang-for-devops-course/oidc-demo/pkg/oidc"
)

// gets token from tokenUrl validating token with jwksUrl and returning token & claims
func getTokenFromCode(tokenUrl, jwksUrl, redirectUri, clientID, clientSecret, code string) ([]*jwt.Token, *jwt.RegisteredClaims, error) {

	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("client_id", clientID)
	values.Add("client_secret", clientSecret)
	values.Add("redirect_uri", redirectUri)
	values.Add("code", code)

	res, err := http.PostForm(tokenUrl, values)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("Statuscode was not 200.")
	}

	var token oidc.Token

	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, nil, fmt.Errorf("Unmarshal token error: %s", err)
	}

	claims := &jwt.RegisteredClaims{}
	parsedIDToken, err := jwt.ParseWithClaims(token.IDToken, claims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"]
		if !ok {
			return nil, fmt.Errorf("kid not found")
		}
		publicKey, err := getPublicKeyFromJwks(jwksUrl, kid.(string))
		if err != nil {
			return nil, fmt.Errorf("getPublicKeyFromJwks error: %s", err)
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Token parsing failed: %s", err)
	}

	AccessTokenClaims := &jwt.RegisteredClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(token.AccessToken, AccessTokenClaims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"]
		if !ok {
			return nil, fmt.Errorf("kid not found")
		}
		publicKey, err := getPublicKeyFromJwks(jwksUrl, kid.(string))
		if err != nil {
			return nil, fmt.Errorf("getPublicKeyFromJwks error: %s", err)
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Token parsing failed: %s", err)
	}

	return []*jwt.Token{parsedIDToken, parsedAccessToken}, claims, nil
}
func getPublicKeyFromJwks(jwksUrl string, kid string) (*rsa.PublicKey, error) {
	res, err := http.Get(jwksUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid statusCode: %d", res.StatusCode)
	}

	// parse jwks
	var jwks oidc.Jwks
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, err
	}

	for _, jwksKeyEntry := range jwks.Keys {
		if jwksKeyEntry.Kid == kid {
			nBytes, err := base64.StdEncoding.DecodeString(jwksKeyEntry.N)
			if err != nil {
				return nil, fmt.Errorf("decodestring error: %s", err)
			}
			n := big.NewInt(0)
			n.SetBytes(nBytes)
			return &rsa.PublicKey{
				N: n,
				E: 65537,
			}, nil
		}
	}
	return nil, fmt.Errorf("No public key found with kid %s", kid)
}
