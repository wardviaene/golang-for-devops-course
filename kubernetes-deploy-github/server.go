package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v45/github"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

type server struct {
	client           *kubernetes.Clientset
	webhookSecretKey []byte
	githubClient     *github.Client
}

func (s server) webhook(w http.ResponseWriter, req *http.Request) {
	payload, err := github.ValidatePayload(req, s.webhookSecretKey)
	if err != nil {
		fmt.Printf("ValidatePayload error: %s", err)
		w.WriteHeader(500)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		fmt.Printf("ParseWebHook error: %s", err)
		w.WriteHeader(500)
		return
	}
	switch event := event.(type) {
	case *github.PushEvent:
		fmt.Printf("Found push event: %+v\n", event)
		fmt.Printf("Files changed/added: %+v\n", getFiles(event.Commits))
		for _, fileName := range getFiles(event.Commits) {
			res, _, err := s.githubClient.Repositories.DownloadContents(context.Background(), *event.Repo.Owner.Name, *event.Repo.Name, fileName, &github.RepositoryContentGetOptions{})
			if err != nil {
				fmt.Printf("DownloadContents error: %s", err)
				w.WriteHeader(500)
				return
			}
			defer res.Close()
			fileBody, err := io.ReadAll(res)
			if err != nil {
				fmt.Printf("ReadAll error: %s", err)
				w.WriteHeader(500)
				return
			}
			_, _, err = s.deploy(context.Background(), fileBody)
			if err != nil {
				fmt.Printf("deploy error: %s", err)
				w.WriteHeader(500)
				return
			}
		}
	default:
		fmt.Printf("Event not found error: %s", event)
		w.WriteHeader(500)
		return
	}
}

func getFiles(commits []*github.HeadCommit) []string {
	files := []string{}
	for _, commit := range commits {
		files = append(files, commit.Added...)
		files = append(files, commit.Modified...)
	}
	uniqueFilesMap := make(map[string]bool)
	for _, file := range files {
		if file != "" { // remove empty elements
			uniqueFilesMap[file] = true
		}
	}
	uniqueFiles := []string{}
	for key := range uniqueFilesMap {
		uniqueFiles = append(uniqueFiles, key)
	}
	return uniqueFiles
}

func (s server) deploy(ctx context.Context, fileBody []byte) (map[string]string, int32, error) {
	var deployment *v1.Deployment

	obj, groupVersionKind, err := scheme.Codecs.UniversalDeserializer().Decode(fileBody, nil, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("Decode error: %s", err)
	}

	switch obj.(type) {
	case *v1.Deployment:
		deployment = obj.(*v1.Deployment)
	default:
		return nil, 0, fmt.Errorf("Unrecognized type: %s\n", groupVersionKind)
	}

	_, err = s.client.AppsV1().Deployments("default").Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		deploymentResponse, err := s.client.AppsV1().Deployments("default").Create(ctx, deployment, metav1.CreateOptions{})
		if err != nil {
			return nil, 0, fmt.Errorf("deployment error: %s", err)
		}
		return deploymentResponse.Spec.Template.Labels, 0, nil
	} else if err != nil && !errors.IsNotFound(err) {
		return nil, 0, fmt.Errorf("deployment get error: %s", err)
	}

	deploymentResponse, err := s.client.AppsV1().Deployments("default").Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("deployment error: %s", err)
	}
	return deploymentResponse.Spec.Template.Labels, *deploymentResponse.Spec.Replicas, nil

}
