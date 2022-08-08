package main

import "fmt"

func main() {
	var f1 float64 = 0.1
	var d1 int64 = 1
	fmt.Printf("result: %v\n", plusOne(f1))
	fmt.Printf("result: %v\n", plusOne(d1))
}

func plusOne[V int64 | float64 | int](value V) V {
	return value + 1
}
