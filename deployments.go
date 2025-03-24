package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentResource adapts Deployment to Resource interface
type DeploymentResource struct {
	*v1.Deployment
}

func (d *DeploymentResource) GetReplicas() *int32  { return d.Spec.Replicas }
func (d *DeploymentResource) SetReplicas(r *int32) { d.Spec.Replicas = r }

func (d *DeploymentResource) GetTemplateAnnotations() map[string]string {
	return d.Spec.Template.Annotations
}

func (d *DeploymentResource) SetTemplateAnnotations(annotations map[string]string) {
	d.Spec.Template.Annotations = annotations
}

// DeploymentUpdater implements ResourceUpdater for Deployments
type DeploymentUpdater struct {
	clientset kubernetes.Interface
}

func (u *DeploymentUpdater) Update(ctx context.Context, namespace string, r Resource) error {
	d, ok := r.(*DeploymentResource)
	if !ok {
		return fmt.Errorf("invalid resource type")
	}
	_, err := u.clientset.AppsV1().Deployments(namespace).Update(ctx, d.Deployment, metav1.UpdateOptions{})
	return err
}

func upscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	resources := make([]Resource, len(deployments.Items))
	for i := range deployments.Items {
		resources[i] = &DeploymentResource{&deployments.Items[i]}
	}

	updater := &DeploymentUpdater{clientset: clientset}
	return upscaleResource(ctx, resources, updater)
}

func downscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	resources := make([]Resource, len(deployments.Items))
	for i := range deployments.Items {
		resources[i] = &DeploymentResource{&deployments.Items[i]}
	}

	updater := &DeploymentUpdater{clientset: clientset}
	return downscaleResource(ctx, resources, updater)
}

func restartDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	resources := make([]Resource, len(deployments.Items))
	for i := range deployments.Items {
		resources[i] = &DeploymentResource{&deployments.Items[i]}
	}

	updater := &DeploymentUpdater{clientset: clientset}
	return restartResource(ctx, resources, updater)
}

func getDeployments(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments: %w", err)
	}
	return deployments, err
}
