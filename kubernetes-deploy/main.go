package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	v1 "k8s.io/api/apps/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var (
		client           *kubernetes.Clientset
		deploymentLabels map[string]string
		err              error
	)
	ctx := context.Background()
	if client, err = getClient(ctx); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	if deploymentLabels, err = deploy(ctx, client); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Deployment successful. Labels: %v\n", deploymentLabels)
	if err = waitForPods(ctx, client, deploymentLabels); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
func getClient(ctx context.Context) (*kubernetes.Clientset, error) {
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
func deploy(ctx context.Context, client *kubernetes.Clientset) (map[string]string, error) {
	appFile, err := ioutil.ReadFile("app.yaml")
	if err != nil {
		return nil, fmt.Errorf("ReadFile error: %s", err)
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode(appFile, nil, nil)
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
	deploymentResponse, err := client.AppsV1().Deployments("default").Apply(ctx, deployment, v1meta.ApplyOptions{
		FieldManager: "go-deploy-fieldmanager",
	})
	if err != nil {
		return nil, err
	}

	return deploymentResponse.Spec.Template.Labels, nil
}

func waitForPods(ctx context.Context, client *kubernetes.Clientset, deploymentLabels map[string]string) error {
	for {
		validatedLabels, err := labels.ValidatedSelectorFromSet(deploymentLabels)
		if err != nil {
			return fmt.Errorf("bad selector set: %v", err)
		}
		pods, err := client.CoreV1().Pods("default").List(ctx, v1meta.ListOptions{
			LabelSelector: validatedLabels.String(),
		})
		if err != nil {
			return fmt.Errorf("Pods list error: %s", err)
		}
		notRunningPods := 0
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Running" {
				notRunningPods++
			}
			fmt.Printf("pods: %s status = %s\n", pod.Name, pod.Status.Phase)
		}

		if notRunningPods == 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}
	return nil
}
