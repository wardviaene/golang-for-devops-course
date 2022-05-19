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

type OccurrenceOutput struct {
	Page  string   `json:"page"`
	Words map[string]int `json:"words"`
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

func (ct *WordsHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "The server is running!")
    return
}

func (ct *WordsHandler) occurrenceHandler(w http.ResponseWriter, r *http.Request) {
  words := make(map[string]int)
  for _, v := range ct.words {
    if _, ok := words[v]; ok {
      words[v]++
    } else {
      words[v] = 1
    }
  }
	occurrenceOutput := OccurrenceOutput{
		Page:  "occurrence",
		Words: words,
	}
	out, err := json.Marshal(occurrenceOutput)
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
	http.HandleFunc("/occurrence", wh.occurrenceHandler)
	http.HandleFunc("/", wh.indexHandler)
	http.ListenAndServe(":8080", nil)
}
