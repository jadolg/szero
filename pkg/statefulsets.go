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
	"time"
)

func UpscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	var resultError error
	upscaledCount := 0
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
				upscaledCount++
			}
		} else {
			log.Infof("Statefulset %s already scaled up", s.Name)
		}
	}
	return upscaledCount, resultError
}

func DownscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList) (int, error) {
	var resultError error
	downscaledCount := 0
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
				downscaledCount++
			}
		} else {
			log.Infof("Statefulset %s already downscaled", s.Name)
		}
	}
	return downscaledCount, resultError
}

func GetStatefulSets(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.StatefulSetList, error) {
	statefulsets, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting statefulsets: %w", err)
	}
	return statefulsets, err
}

func IsStatefulSetReady(ss *v1.StatefulSet, downscaled bool) bool {
	if downscaled {
		return ss.Status.Replicas == 0 && ss.Status.ReadyReplicas == 0
	}
	return ss.Status.AvailableReplicas == *ss.Spec.Replicas
}

func WaitForStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList, timeout time.Duration, downscaled bool) error {
	ticker := time.NewTicker(1 * time.Second)
	timeoutAfter := time.After(timeout)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timeout waiting for statefulsets to reconcile")
		case <-ticker.C:
			done := true
			for _, d := range statefulsets.Items {
				ss, err := clientset.AppsV1().StatefulSets(d.Namespace).Get(ctx, d.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("error getting statefulset %s: %w", d.Name, err)
				}
				if IsStatefulSetReady(ss, downscaled) {
					continue
				} else {
					done = false
					break
				}
			}
			if done {
				return nil
			}
		}
	}
}
