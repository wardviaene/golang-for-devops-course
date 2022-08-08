package main

import (
	"fmt"
	"reflect"
)

func main() {
	var t1 string = "this is a string"
	var t2 *string = &t1
	discoverType(t2)
	var t3 int = 123
	discoverType(t3)
	discoverType(nil)
}

func discoverType(t any) {
	switch v := t.(type) {
	case string:
		t2 := v + "..."
		fmt.Printf("String found: %s\n", t2)
	case *string:
		fmt.Printf("Pointer string found: %s\n", *v)
	case int:
		fmt.Printf("We have an integer: %d\n", v)
	default:
		myType := reflect.TypeOf(t)
		if myType == nil {
			fmt.Printf("type is nil\n")
		} else {
			fmt.Printf("Type not found: %s\n", myType)
		}

	}
}
