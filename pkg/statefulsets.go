package pkg

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

func UpscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	var resultError error
	upscaledStatefulsets := 0
	for _, s := range statefulsets.Items {
		replicas, downscaled := s.Annotations[replicasAnnotation]
		if downscaled {
			intReplicas, err := strconv.ParseInt(replicas, 10, 32)
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error converting replicas to int: %w", err), resultError)
				continue
			}
			log.Infof("Scaling up statefulset %s to %d replicas", s.Name, intReplicas)
			*s.Spec.Replicas = int32(intReplicas)
			delete(s.Annotations, replicasAnnotation)
			_, err = clientset.AppsV1().StatefulSets(s.Namespace).Update(ctx, &s, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling up statefulset %s: %v", s.Name, err), resultError)
			} else {
				upscaledStatefulsets++
			}
		} else {
			log.Infof("Statefulset %s already scaled up", s.Name)
		}
	}
	return upscaledStatefulsets, resultError
}

func DownscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	var resultError error
	downscaledStatefulsets := 0
	for _, s := range statefulsets.Items {
		_, downscaled := s.Annotations[replicasAnnotation]
		if !downscaled || *s.Spec.Replicas > 0 {
			log.Infof("Scaling down statefulset %s from %d replicas", s.Name, s.Status.Replicas)
			if !downscaled {
				s.Annotations[replicasAnnotation] = fmt.Sprintf("%d", *s.Spec.Replicas)
			}
			*s.Spec.Replicas = 0
			_, err := clientset.AppsV1().StatefulSets(s.Namespace).Update(ctx, &s, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling down statefulset %s: %v", s.Name, err), resultError)
			} else {
				downscaledStatefulsets++
			}
		} else {
			log.Infof("Statefulset %s already downscaled", s.Name)
		}
	}
	return downscaledStatefulsets, resultError
}

func GetStatefulSets(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.StatefulSetList, error) {
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting statefulsets: %w", err)
	}
	return statefulsets, err
}
