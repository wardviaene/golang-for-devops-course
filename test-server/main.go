package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WordsOutput struct {
	Page  string   `json:"page"`
	Input string   `json:"input"`
	Words []string `json:"words"`
}

type WordsHandler struct {
	words []string
}

func (ct *WordsHandler) wordsHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input != "" {
		ct.words = append(ct.words, input)
	}

	wordsOutput := WordsOutput{
		Page:  "words",
		Input: input,
		Words: ct.words,
	}
	out, err := json.Marshal(wordsOutput)
	if err != nil {
		fmt.Fprintf(w, "marshal error")
		return
	}
	fmt.Fprint(w, string(out))
}

func main() {
	wh := &WordsHandler{
		words: []string{},
	}
	http.HandleFunc("/words", wh.wordsHandler)
	http.ListenAndServe(":8080", nil)
}