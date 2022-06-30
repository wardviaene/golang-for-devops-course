package oidc

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
)

func GetTokenFromCode(tokenUrl, jwksUrl, clientID, clientSecret, code string) (*jwt.Token, *jwt.StandardClaims, error) {
	form := url.Values{}
	form.Add("grant_type", "authorization_code")
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("code", code)

	res, err := http.PostForm(tokenUrl, form)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode != 200 {
		return nil, nil, fmt.Errorf("StatusCode %d, error: %s", res.StatusCode, body)
	}
	var tokenReply Token
	err = json.Unmarshal(body, &tokenReply)
	if err != nil {
		return nil, nil, err
	}
	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenReply.IDToken, claims, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"]
		if !ok {
			return nil, fmt.Errorf("kid not found")
		}

		rsaPublicKey, err := getPublicKeyFromJwks(jwksUrl, kid.(string))
		if err != nil {
			return nil, err
		}
		return rsaPublicKey, nil
	})
	return token, claims, nil
}

func getPublicKeyFromJwks(jwksUrl, kid string) (*rsa.PublicKey, error) {
	res, err := http.Get(jwksUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var jwks Jwks
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, err
	}
	for _, keys := range jwks.Keys {
		if keys.Kid == kid {
			nDecoded, err := base64.StdEncoding.DecodeString(keys.N)
			if err != nil {
				return nil, err
			}
			n := big.NewInt(0)
			n.SetBytes(nDecoded)

			return &rsa.PublicKey{
				N: n,
			}, nil
		}
	}
	return nil, fmt.Errorf("Coudln't find public key at jwks url")
}
