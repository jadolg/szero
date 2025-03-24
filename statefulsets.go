package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSetResource adapts StatefulSet to Resource interface
type StatefulSetResource struct {
	*v1.StatefulSet
}

func (s *StatefulSetResource) GetReplicas() *int32  { return s.Spec.Replicas }
func (s *StatefulSetResource) SetReplicas(r *int32) { s.Spec.Replicas = r }
func (s *StatefulSetResource) GetTemplateAnnotations() map[string]string {
	return s.Spec.Template.Annotations
}

func (s *StatefulSetResource) SetTemplateAnnotations(annotations map[string]string) {
	s.Spec.Template.Annotations = annotations
}

// StatefulSetUpdater implements ResourceUpdater for Statefulsets
type StatefulSetUpdater struct {
	clientset kubernetes.Interface
}

func (u *StatefulSetUpdater) Update(ctx context.Context, namespace string, r Resource) error {
	d, ok := r.(*StatefulSetResource)
	if !ok {
		return fmt.Errorf("invalid resource type")
	}
	_, err := u.clientset.AppsV1().StatefulSets(namespace).Update(ctx, d.StatefulSet, metav1.UpdateOptions{})
	return err
}

func upscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	resources := make([]Resource, len(statefulsets.Items))
	for i := range statefulsets.Items {
		resources[i] = &StatefulSetResource{&statefulsets.Items[i]}
	}

	updater := &StatefulSetUpdater{clientset: clientset}
	return upscaleResource(ctx, resources, updater)
}

func downscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	resources := make([]Resource, len(statefulsets.Items))
	for i := range statefulsets.Items {
		resources[i] = &StatefulSetResource{&statefulsets.Items[i]}
	}

	updater := &StatefulSetUpdater{clientset: clientset}
	return downscaleResource(ctx, resources, updater)
}

func restartStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	resources := make([]Resource, len(statefulsets.Items))
	for i := range statefulsets.Items {
		resources[i] = &StatefulSetResource{&statefulsets.Items[i]}
	}

	updater := &StatefulSetUpdater{clientset: clientset}
	return restartResource(ctx, resources, updater)
}

func getStatefulSets(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.StatefulSetList, error) {
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting statefulsets: %w", err)
	}
	return statefulsets, err
}
