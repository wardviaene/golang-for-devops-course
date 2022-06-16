package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type assignment1 struct {
	Page         string             `json:"page"`
	Words        []string           `json:"words"`
	Percentages  map[string]float64 `json:"percentages"`
	Special      []*string          `json:"special"`
	ExtraSpecial []any              `json:"extraSpecial"`
}

func (ct *WordsHandler) assignment1(w http.ResponseWriter, r *http.Request) {
	one := "one"
	two := "two"
	words := []string{"one", "two", "three", "four", "five", "six", "seven", "eigth", "nine", "ten"}
	numbers := []float64{0.33, 0.66, 0.1, 0, 1, 0.99, 0.88, 0.5, 0.1, 0.2}
	rand.Seed(time.Now().UnixNano())
	percentages := make(map[string]float64)
	wordsRand := make([]string, 5)
	for i := 0; i < 5; i++ {
		randomInt := rand.Intn(9)
		wordsRand[i] = words[randomInt]
		percentages[words[randomInt]] = numbers[randomInt]
	}
	wordsOutput := assignment1{
		Page:         "assignment1",
		Words:        wordsRand,
		Percentages:  percentages,
		Special:      []*string{&one, &two, nil},
		ExtraSpecial: []any{1, 2, "3"},
	}
	out, err := json.Marshal(wordsOutput)
	if err != nil {
		fmt.Fprintf(w, "marshal error")
		return
	}
	fmt.Fprint(w, string(out))
}
