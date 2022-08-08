package main

import "fmt"

func main() {
	var t1 int64 = 123
	var t2 *int64 = &t1
	discoverType(t2)
}

func discoverType(t any) {
	switch v := t.(type) {
	case string:
		fmt.Printf("This is a string: %s", v)
	case int64:
		fmt.Printf("This is an int64: %d", v)
	case *int64:
		fmt.Printf("This is an int64 pointer: %v", v)
	default:
		fmt.Printf("Could not recognize type")
	}
}
