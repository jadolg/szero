package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getClientset(kubeconfig string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Error building config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error building clientset: %v", err)
	}

	return clientset
}

func main() {
	var kubeconfig *string
	namespace := flag.String("namespace", "default", "the namespace to use")
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	deploymentScales := make(map[string]int32)

	clientset := getClientset(*kubeconfig)
	ctx := context.Background()
	err := downscaleDeployments(clientset, namespace, ctx, deploymentScales)
	if err != nil {
		log.Fatalf("Error downscaling deployments %v", err)
	}

	fmt.Print("Press 'Enter' to start scaling up again...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	downscaledDeployments, err := clientset.AppsV1().Deployments(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting deployments %v", err)
	}

	upscaleDeployments(downscaledDeployments, deploymentScales, clientset, namespace, ctx)
}

func upscaleDeployments(downscaledDeployments *v1.DeploymentList, deploymentScales map[string]int32, clientset *kubernetes.Clientset, namespace *string, ctx context.Context) {
	for _, d := range downscaledDeployments.Items {
		fmt.Printf("Scaling up deployment %s\n", d.Name)
		*d.Spec.Replicas = deploymentScales[d.Name]
		_, err := clientset.AppsV1().Deployments(*namespace).Update(ctx, &d, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Error scaling up deployment %s: %v\n", d.Name, err)
		}
	}

	deployments, err := clientset.AppsV1().Deployments(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting deployments %v", err)
	}
	printDeployments(deployments)
}

func downscaleDeployments(clientset *kubernetes.Clientset, namespace *string, ctx context.Context, deploymentScales map[string]int32) error {
	fmt.Printf("Scaling deployments down in namespace: %s\n", *namespace)
	deployments, err := clientset.AppsV1().Deployments(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting deployments %v", err)
	}
	printDeployments(deployments)
	for _, d := range deployments.Items {
		fmt.Printf("Scaling down deployment %s\n", d.Name)
		deploymentScales[d.Name] = *d.Spec.Replicas
		*d.Spec.Replicas = 0
		_, err := clientset.AppsV1().Deployments(*namespace).Update(ctx, &d, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("Error scaling down deployment %s: %v\n", d.Name, err)
		}
	}
	return err
}

func printDeployments(deployments *v1.DeploymentList) {
	fmt.Printf("[DEPLOYMENTS] -- %d\n", len(deployments.Items))
	for _, d := range deployments.Items {
		fmt.Printf("\t- %s (%d/%d)\n", d.Name, d.Status.Replicas, *d.Spec.Replicas)
	}
}
