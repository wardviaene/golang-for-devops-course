package main

import (
	"fmt"
	"io"
	"log"
)

type MySlowReader struct {
	contents string
	pos      int
}

func (m *MySlowReader) Read(p []byte) (int, error) {
	if m.pos+1 <= len(m.contents) {
		n := copy(p, m.contents[m.pos:m.pos+1])
		m.pos++
		return n, nil
	}
	return 0, io.EOF
}

func main() {
	myReaderInstance := &MySlowReader{
		contents: "a",
	}
	out, err := io.ReadAll(myReaderInstance)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("output: %s", out)
}
