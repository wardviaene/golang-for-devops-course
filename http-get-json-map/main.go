package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Page struct {
	Name string `json:"page"`
}

type Words struct {
	Input string   `json:"input"`
	Words []string `json:"words"`
}

type Occurrence struct {
	Words map[string]int `json:"words"`
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Printf("Usage: ./http-get <url>\n")
		os.Exit(1)
	}

	if _, err := url.ParseRequestURI(args[1]); err != nil {
		fmt.Printf("Usage: ./http-get <url>\n\nURL is not valid URL: %s\n", args[1])
		os.Exit(1)
	}

	response, err := http.Get(args[1])

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		log.Fatal(err)
	}

	if response.StatusCode != 200 {
		fmt.Printf("Invalid output (HTTP Code %d): %s\n", response.StatusCode, string(body))
		os.Exit(1)
	}

	var page Page

	err = json.Unmarshal(body, &page)
	if err != nil {
		log.Fatal(err)
	}

	switch page.Name {
	case "words":
		var words Words
		err = json.Unmarshal(body, &words)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("JSON: Parsed:\nPage: %s\nWords: %s\n", page.Name, strings.Join(words.Words, ", "))
	case "occurrence":
		var occurrence Occurrence
		err = json.Unmarshal(body, &occurrence)
		if err != nil {
			log.Fatal(err)
		}

		if val, ok := occurrence.Words["word5"]; ok {
			fmt.Printf("Found word1: %d\n", val)
		}

		for word, occurrence := range occurrence.Words {
			fmt.Printf("%s: %d\n", word, occurrence)
		}

	default:
		fmt.Printf("Page not found\n")
	}

}
