package pkg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func UpscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList, dryRun bool) ([]ScaleInfo, error) {
	var resultError error
	var results []ScaleInfo
	for _, s := range statefulsets.Items {
		upscaled, replicas, err := upscaleStatefulset(ctx, clientset, s.Namespace, s.Name, dryRun)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling up statefulset %s: %w", s.Name, err), resultError)
		}
		info := ScaleInfo{
			Name:     s.Name,
			Replicas: replicas,
			Scaled:   upscaled,
		}
		if !upscaled {
			info.Warning = "already scaled up"
		}
		results = append(results, info)
	}
	return results, resultError
}

func DownscaleStatefulSets(ctx context.Context, clientset kubernetes.Interface, statefulsets *v1.StatefulSetList, dryRun bool) ([]ScaleInfo, error) {
	var resultError error
	var results []ScaleInfo
	for _, s := range statefulsets.Items {
		downscaled, originalReplicas, err := downscaleStatefulset(ctx, clientset, s.Namespace, s.Name, dryRun)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling down statefulset %s: %w", s.Name, err), resultError)
		}
		info := ScaleInfo{
			Name:     s.Name,
			Replicas: originalReplicas,
			Scaled:   downscaled,
		}
		if !downscaled {
			info.Warning = "already downscaled"
		}
		results = append(results, info)
	}
	return results, resultError
}

func upscaleStatefulset(ctx context.Context, clientset kubernetes.Interface, namespace string, name string, dryRun bool) (bool, int32, error) {
	var targetReplicas int32
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		s, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		replicas, downscaled := s.Annotations[replicasAnnotation]
		if downscaled {
			intReplicas, err := strconv.ParseInt(replicas, 10, 32)
			if err != nil {
				return fmt.Errorf("error converting replicas to int: %w", err)
			}
			targetReplicas = int32(intReplicas)
			if dryRun {
				w = true
				return nil
			}
			*s.Spec.Replicas = int32(intReplicas)
			delete(s.Annotations, replicasAnnotation)
			_, err = clientset.AppsV1().StatefulSets(s.Namespace).Update(ctx, s, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		}

		return nil
	})
	return w, targetReplicas, err
}

func downscaleStatefulset(ctx context.Context, clientset kubernetes.Interface, namespace string, name string, dryRun bool) (bool, int32, error) {
	var originalReplicas int32
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		s, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		_, downscaled := s.Annotations[replicasAnnotation]
		if !downscaled || *s.Spec.Replicas > 0 {
			originalReplicas = *s.Spec.Replicas
			if dryRun {
				w = true
				return nil
			}
			if !downscaled {
				s.Annotations[replicasAnnotation] = fmt.Sprintf("%d", *s.Spec.Replicas)
			}
			*s.Spec.Replicas = 0
			_, err := clientset.AppsV1().StatefulSets(s.Namespace).Update(ctx, s, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		}
		return nil
	})
	return w, originalReplicas, err
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
