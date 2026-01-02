package pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func GetDaemonsets(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DaemonSetList, error) {
	daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonsets, nil
}

func DownscaleDaemonsets(ctx context.Context, clientset kubernetes.Interface, daemonsets *v1.DaemonSetList, dryRun bool) ([]ScaleInfo, error) {
	var resultError error
	var results []ScaleInfo
	for _, d := range daemonsets.Items {
		downscaled, err := downscaleDaemonset(ctx, clientset, d.Namespace, d.Name, dryRun)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling down resource %s: %w", d.GetName(), err), resultError)
		}
		info := ScaleInfo{
			Name:     d.Name,
			Replicas: 0, // DaemonSets don't have replicas
			Scaled:   downscaled,
		}
		if !downscaled {
			info.Warning = "already downscaled"
		}
		results = append(results, info)
	}
	return results, resultError
}

func UpscaleDaemonsets(ctx context.Context, clientset kubernetes.Interface, daemonsets *v1.DaemonSetList, dryRun bool) ([]ScaleInfo, error) {
	var resultError error
	var results []ScaleInfo
	for _, d := range daemonsets.Items {
		upscaled, err := upscaleDaemonset(ctx, clientset, d.Namespace, d.Name, dryRun)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling up resource %s: %w", d.GetName(), err), resultError)
		}
		info := ScaleInfo{
			Name:     d.Name,
			Replicas: 0, // DaemonSets don't have replicas
			Scaled:   upscaled,
		}
		if !upscaled {
			info.Warning = "already scaled up"
		}
		results = append(results, info)
	}
	return results, resultError
}

func downscaleDaemonset(ctx context.Context, clientset kubernetes.Interface, namespace string, name string, dryRun bool) (bool, error) {
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if _, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]; !exists {
			if dryRun {
				w = true
				return nil
			}
			if d.Spec.Template.Spec.NodeSelector == nil {
				d.Spec.Template.Spec.NodeSelector = make(map[string]string)
			}
			d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation] = "true"
			_, err := clientset.AppsV1().DaemonSets(d.Namespace).Update(ctx, d, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		}
		return nil
	})
	return w, err
}

func upscaleDaemonset(ctx context.Context, clientset kubernetes.Interface, namespace string, name string, dryRun bool) (bool, error) {
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if _, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]; exists {
			if dryRun {
				w = true
				return nil
			}
			delete(d.Spec.Template.Spec.NodeSelector, noscheduleAnnotation)
			_, err := clientset.AppsV1().DaemonSets(d.Namespace).Update(ctx, d, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		}

		return nil
	})
	return w, err
}

func IsDaemonSetReady(ds *v1.DaemonSet, downscaled bool) bool {
	if downscaled {
		return ds.Status.NumberReady == 0 && ds.Status.NumberMisscheduled == 0
	}
	return ds.Status.NumberReady == ds.Status.DesiredNumberScheduled
}

func WaitForDaemonSets(ctx context.Context, clientset kubernetes.Interface, daemonsets *v1.DaemonSetList, timeout time.Duration, downscaled bool) error {
	ticker := time.NewTicker(1 * time.Second)
	timeoutAfter := time.After(timeout)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timeout waiting for DaemonSets to reconcile")
		case <-ticker.C:
			done := true
			for _, d := range daemonsets.Items {
				ds, err := clientset.AppsV1().DaemonSets(d.Namespace).Get(ctx, d.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("error getting DaemonSet %s: %w", d.Name, err)
				}
				if IsDaemonSetReady(ds, downscaled) {
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
