package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v45/github"
	"k8s.io/client-go/kubernetes"
)

type server struct {
	client           *kubernetes.Clientset
	githubClient     *github.Client
	webhookSecretKey string
}

func (s server) webhook(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	payload, err := github.ValidatePayload(req, []byte(s.webhookSecretKey))
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("ValidatePayload error: %s", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("ValidatePayload error: %s", err)
		return
	}
	switch event := event.(type) {
	case *github.Hook:
		fmt.Printf("Hook is created\n")
	case *github.PushEvent:
		files := getFiles(event.Commits)
		fmt.Printf("Found files: %s\n", strings.Join(files, ", "))
		for _, filename := range files {
			downloadedFile, _, err := s.githubClient.Repositories.DownloadContents(ctx, *event.Repo.Owner.Name, *event.Repo.Name, filename, &github.RepositoryContentGetOptions{})
			if err != nil {
				w.WriteHeader(500)
				fmt.Printf("DownloadContents error: %s", err)
				return
			}
			defer downloadedFile.Close()
			fileBody, err := io.ReadAll(downloadedFile)
			if err != nil {
				w.WriteHeader(500)
				fmt.Printf("ReadAll error: %s", err)
				return
			}
			_, _, err = deploy(ctx, s.client, fileBody)
			if err != nil {
				w.WriteHeader(500)
				fmt.Printf("deploy error: %s", err)
				return
			}
			fmt.Printf("Deploy of %s finished\n", filename)
		}
	default:
		w.WriteHeader(500)
		fmt.Printf("Event not found: %s", event)
		return
	}
}

func getFiles(commits []*github.HeadCommit) []string {
	allFiles := []string{}
	for _, commit := range commits {
		allFiles = append(allFiles, commit.Added...)
		allFiles = append(allFiles, commit.Modified...)
	}
	allUniqueFiles := make(map[string]bool)
	for _, filename := range allFiles {
		allUniqueFiles[filename] = true
	}
	allUniqueFilesSlice := []string{}
	for filename := range allUniqueFiles {
		allUniqueFilesSlice = append(allUniqueFilesSlice, filename)
	}
	return allUniqueFilesSlice
}
