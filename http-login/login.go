package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LoginRequest struct {
	Password string `json:"password"`
}
type LoginResponse struct {
	Token string `json:"token"`
}

func doLoginRequest(client http.Client, requestURL, password string) (string, error) {
	loginRequest := LoginRequest{
		Password: password,
	}

	body, err := json.Marshal(loginRequest)

	response, err := client.Post(requestURL, "application/json", bytes.NewBuffer(body))

	if err != nil {
		return "", fmt.Errorf("Get error: %s", err)
	}

	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)

	var loginResponse LoginResponse

	err = json.Unmarshal(res, &loginResponse)
	if err != nil {
		return "", RequestError{
			Err:      fmt.Sprintf("Page unmarshal error: %s", err),
			HTTPCode: response.StatusCode,
			Body:     string(body),
		}
	}

	return loginResponse.Token, nil
}
