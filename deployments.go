package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const replicasAnnotation = "szero/replicas"
const restartedAtAnnotation = "kubernetes.io/restartedAt"
const changeCauseAnnotation = "kubernetes.io/change-cause"

func getClientset(kubeconfig, context string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error building config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error building clientset: %w", err)
	}

	return clientset, nil
}

func upscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	var resultError error
	upscaledDeployments := 0
	for _, d := range deployments.Items {
		replicas, downscaled := d.Annotations[replicasAnnotation]
		if downscaled {
			intReplicas, err := strconv.ParseInt(replicas, 10, 32)
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error converting replicas to int: %w", err))
				continue
			}
			log.Infof("Scaling up deployment %s to %d replicas", d.Name, intReplicas)
			*d.Spec.Replicas = int32(intReplicas)
			delete(d.Annotations, replicasAnnotation)
			_, err = clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling up deployment %s: %v", d.Name, err))
			} else {
				upscaledDeployments++
			}
		} else {
			log.Infof("Deployment %s already scaled up", d.Name)
		}
	}
	return upscaledDeployments, resultError
}

func downscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	var resultError error
	downscaledDeployments := 0
	for _, d := range deployments.Items {
		_, downscaled := d.Annotations[replicasAnnotation]
		if !downscaled || *d.Spec.Replicas > 0 {
			log.Infof("Scaling down deployment %s from %d replicas", d.Name, d.Status.Replicas)
			if !downscaled {
				d.Annotations[replicasAnnotation] = fmt.Sprintf("%d", *d.Spec.Replicas)
			}
			*d.Spec.Replicas = 0
			_, err := clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling down deployment %s: %v", d.Name, err))
			} else {
				downscaledDeployments++
			}
		} else {
			log.Infof("Deployment %s already downscaled", d.Name)
		}
	}
	return downscaledDeployments, resultError
}

func restartDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	var resultError error
	restartedDeployments := 0
	for _, d := range deployments.Items {
		log.Infof("Restarting deployment %s", d.Name)
		if d.Spec.Template.Annotations == nil {
			d.Spec.Template.Annotations = make(map[string]string)
		}
		d.Spec.Template.Annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
		d.Spec.Template.Annotations[changeCauseAnnotation] = "Restarted by szero"
		_, err := clientset.AppsV1().Deployments(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error restarting deployment %s: %v", d.Name, err))
		} else {
			restartedDeployments++
		}
	}
	return restartedDeployments, resultError
}

func getDeployments(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments: %w", err)
	}
	return deployments, err
}
