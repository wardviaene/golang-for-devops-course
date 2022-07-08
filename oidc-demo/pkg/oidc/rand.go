package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

func GetRandomString(n int) (string, error) {
	buf := make([]byte, n)

	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", fmt.Errorf("crypto/rand Reader error: %s", err)
	}

	randomStr := base64.URLEncoding.EncodeToString(buf)
	randomStr = strings.Replace(randomStr, "=", "", -1)

	return randomStr, nil
}
