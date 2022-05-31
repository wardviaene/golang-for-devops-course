package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		client *kubernetes.Clientset
		err    error
	)
	ctx := context.Background()
	if client, err = getClient(ctx, false); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	serverInstance := server{
		client:           client,
		webhookSecretKey: []byte(os.Getenv("WEBHOOK_SECRET")),
		githubClient:     getGitHubClient(ctx, os.Getenv("GITHUB_TOKEN")),
	}

	http.HandleFunc("/webhook", serverInstance.webhook)

	err = http.ListenAndServe(":8090", nil)
	fmt.Printf("Exited: %s\n", err)
}
func getClient(ctx context.Context, inCluster bool) (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	if inCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, err
		}
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getGitHubClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
