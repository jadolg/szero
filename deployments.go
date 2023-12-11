package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const ReplicasAnnotation = "szero/replicas"

func getClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error building clientset: %w", err)
	}

	return clientset, nil
}

func upscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) error {
	var resultError error
	for _, d := range deployments.Items {
		replicas, downscaled := d.Annotations[ReplicasAnnotation]

		intReplicas, err := strconv.Atoi(replicas)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error converting replicas to int: %w", err))
			continue
		}

		if downscaled {
			log.Infof("Scaling up deployment %s", d.Name)
			*d.Spec.Replicas = int32(intReplicas)
			delete(d.Annotations, ReplicasAnnotation)
			_, err := clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling up deployment %s: %v", d.Name, err))
			}
		} else {
			log.Debugf("Deployment %s already scaled up", d.Name)
		}
	}
	return resultError
}

func downscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) error {
	var resultError error
	for _, d := range deployments.Items {
		_, downscaled := d.Annotations[ReplicasAnnotation]
		if !downscaled {
			log.Infof("Scaling down deployment %s", d.Name)
			d.Annotations[ReplicasAnnotation] = fmt.Sprintf("%d", *d.Spec.Replicas)
			*d.Spec.Replicas = 0
			_, err := clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling down deployment %s: %v", d.Name, err))
			}
		} else {
			log.Debugf("Deployment %s already downscaled", d.Name)
		}
	}
	return resultError
}

func getDeployments(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments: %w", err)
	}
	return deployments, err
}
