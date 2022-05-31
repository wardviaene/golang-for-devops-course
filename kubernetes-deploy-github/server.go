package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v45/github"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
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
			s.deploy(context.Background(), fileBody)
		}
	default:
		fmt.Printf("Event not found error: %s", err)
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
		uniqueFilesMap[file] = true
	}
	uniqueFiles := make([]string, len(uniqueFilesMap))
	for key := range uniqueFilesMap {
		uniqueFiles = append(uniqueFiles, key)
	}
	return uniqueFiles
}

func (s server) deploy(ctx context.Context, fileBody []byte) (map[string]string, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode(fileBody, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error: %s", err)
	}
	var deployment *appsv1.DeploymentApplyConfiguration
	switch obj.(type) {
	case *v1.Deployment:
		deployment, err = appsv1.ExtractDeployment(obj.(*v1.Deployment), "go-deploy-fieldmanager")
		if err != nil {
			return nil, fmt.Errorf("ExtractDeployment error: %s", err)
		}
	default:
		return nil, fmt.Errorf("type not found: %s", groupVersionKind.Kind)
	}
	deploymentResponse, err := s.client.AppsV1().Deployments("default").Apply(ctx, deployment, metav1.ApplyOptions{
		FieldManager: "go-deploy-fieldmanager",
	})
	if err != nil {
		return nil, err
	}

	return deploymentResponse.Spec.Template.Labels, nil
}
