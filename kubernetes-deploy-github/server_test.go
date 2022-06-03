package main

import (
	"strings"
	"testing"

	"github.com/google/go-github/v45/github"
)

func TestGetFiles(t *testing.T) {
	files := getFiles([]*github.HeadCommit{
		{
			Added: []string{"test.txt"},
		},
	})
	if len(files) != 1 {
		t.Fatalf("len files is not 1. Got: %v", strings.Join(files, ","))
	}
	if files[0] != "test.txt" {
		t.Fatalf("first element is not test.txt. Got: %s", files[0])
	}
}
