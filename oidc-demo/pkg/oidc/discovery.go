package oidc

import (
	"encoding/json"
	"io"
	"net/http"
)

func ParseDiscovery(url string) (Discovery, error) {
	var discovery Discovery
	res, err := http.Get(url)
	if err != nil {
		return discovery, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err = json.Unmarshal(body, &discovery); err != nil {
		return discovery, err
	}
	return discovery, nil
}
