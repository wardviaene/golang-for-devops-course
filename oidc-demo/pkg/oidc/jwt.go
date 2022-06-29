package oidc

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
)

func GetToken(tokenUrl, jwksUrl, clientID, clientSecret, redirectUri string) (*jwt.Token, error) {
	res, err := http.Get(fmt.Sprintf("%s?grant_type=authorization_code&client_id=%s&client_secret=%s&redirect_uri=%s", tokenUrl, clientID, clientSecret, redirectUri))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var tokenReply Token
	err = json.Unmarshal(body, &tokenReply)
	if err != nil {
		return nil, err
	}
	claims := jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenReply.IDToken, &claims, func(token *jwt.Token) (interface{}, error) {
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
	return token, nil
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
