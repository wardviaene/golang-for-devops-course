package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v1 "k8s.io/api/apps/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	ctx := context.Background()
	if err := deploy(ctx); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
func deploy(ctx context.Context) error {
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	var deployment v1.Deployment
	//yaml.Unmarshal()
	_, err = client.AppsV1().Deployments("default").Create(ctx, &deployment, v1meta.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}
