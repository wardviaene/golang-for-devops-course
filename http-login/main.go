package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Page struct {
	Name string `json:"page"`
}

type Response interface {
	GetResponse() string
}

type Words struct {
	Input string   `json:"input"`
	Words []string `json:"words"`
}

func (w Words) GetResponse() string {
	return fmt.Sprintf("Words: %s", strings.Join(w.Words, ", "))
}

type Occurrence struct {
	Words map[string]int `json:"words"`
}

func (o Occurrence) GetResponse() string {
	words := []string{}
	for word, occurrence := range o.Words {
		words = append(words, fmt.Sprintf("%s (%d)", word, occurrence))
	}
	return fmt.Sprintf("Words: %s", strings.Join(words, ", "))
}

func main() {
	var (
		requestURL string
		password   string
		parsedURL  *url.URL
		err        error
	)
	flag.StringVar(&requestURL, "url", "", "url to access")
	flag.StringVar(&password, "password", "", "use a password to access our api")

	flag.Parse()

	if parsedURL, err = url.ParseRequestURI(requestURL); err != nil {
		fmt.Printf("Help: ./http-get -h\nURL is not valid URL: %s\n", requestURL)
		os.Exit(1)
	}

	client := http.Client{}

	if password != "" {
		token, err := doLoginRequest(client, parsedURL.Scheme+"://"+parsedURL.Host+"/login", password)
		if err != nil {
			if requestErr, ok := err.(RequestError); ok {
				fmt.Printf("Login failed: %s (HTTP Error: %d, Body: %s)\n", requestErr.Error(), requestErr.HTTPCode, requestErr.Body)
				os.Exit(1)
			}
			fmt.Printf("Login failed: %s\n", err)
			os.Exit(1)
		}
		client.Transport = MyJWTTransport{
			transport: http.DefaultTransport,
			token:     token,
		}
	}

	res, err := doRequest(client, parsedURL.String())
	if err != nil {
		if requestErr, ok := err.(RequestError); ok {
			fmt.Printf("Error occurred: %s (HTTP Error: %d, Body: %s)\n", requestErr.Error(), requestErr.HTTPCode, requestErr.Body)
			os.Exit(1)
		}
		fmt.Printf("Error occurred: %s\n", err)
		os.Exit(1)
	}
	if res == nil {
		fmt.Printf("No response\n")
		os.Exit(1)
	}
	fmt.Printf("Response: %s\n", res.GetResponse())
}

func doRequest(client http.Client, requestURL string) (Response, error) {

	response, err := client.Get(requestURL)

	if err != nil {
		return nil, fmt.Errorf("Get error: %s", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("ReadAll error: %s", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid output (HTTP Code %d): %s\n", response.StatusCode, string(body))
	}

	var page Page

	if !json.Valid(body) {
		return nil, RequestError{
			Err:      fmt.Sprintf("Response is not a json"),
			HTTPCode: response.StatusCode,
			Body:     string(body),
		}
	}

	err = json.Unmarshal(body, &page)
	if err != nil {
		return nil, RequestError{
			Err:      fmt.Sprintf("Page unmarshal error: %s", err),
			HTTPCode: response.StatusCode,
			Body:     string(body),
		}
	}

	switch page.Name {
	case "words":
		var words Words
		err = json.Unmarshal(body, &words)
		if err != nil {
			return nil, fmt.Errorf("Words unmarshal error: %s", err)
		}

		return words, nil
	case "occurrence":
		var occurrence Occurrence
		err = json.Unmarshal(body, &occurrence)
		if err != nil {
			return nil, fmt.Errorf("Occurrence unmarshal error: %s", err)
		}

		return occurrence, nil
	}

	return nil, nil
}
