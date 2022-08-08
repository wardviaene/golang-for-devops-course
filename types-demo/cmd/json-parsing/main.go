package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
)

type MyJson struct {
	Test  any    `json:"test"`
	Test3 string `json:"test3"`
}

func main() {
	var jsonParsed MyJson
	err := json.Unmarshal([]byte(`{"test": { "test2": [1,2,3] }, "test3": "..." }`), &jsonParsed)
	if err != nil {
		log.Fatal(err)
	}
	switch v := jsonParsed.Test.(type) {
	case map[string]any:
		fmt.Printf("Map found: %v\n", v)
		field1, ok := v["test2"]
		if ok {
			switch v2 := field1.(type) {
			case []any:
				fmt.Printf("I found a []any\n")
				for _, v2Element := range v2 {
					fmt.Printf("Type: %s\n", reflect.TypeOf(v2Element))
					if reflect.TypeOf(v2Element).String() == "float64" {
						fmt.Printf("Int: %d\n", int(v2Element.(float64)))
					} else {
						fmt.Printf("Didn't recognize v2Element\n")
					}
				}
			default:
				fmt.Printf("Type not found: %s\n", reflect.TypeOf(v2))
			}
		}
	default:
		fmt.Printf("Type not found: %s\n", reflect.TypeOf(jsonParsed))
	}
}
