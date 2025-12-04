package pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetDaemonsets(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DaemonSetList, error) {
	daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return daemonsets, nil
}

func DownscaleDaemonsets(ctx context.Context, clientset kubernetes.Interface, daemonsets *v1.DaemonSetList) (int, error) {
	downscaledCount := 0
	var resultError error
	for _, d := range daemonsets.Items {
		if _, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]; !exists {
			log.Infof("Scaling down daemonset %s", d.Name)
			if d.Spec.Template.Spec.NodeSelector == nil {
				d.Spec.Template.Spec.NodeSelector = make(map[string]string)
			}
			d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation] = "true"
			_, err := clientset.AppsV1().DaemonSets(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling down resource %s: %w", d.GetName(), err), resultError)
			} else {
				downscaledCount++
			}
		} else {
			log.Infof("Daemonset %s already downscaled", d.Name)
		}
	}
	return downscaledCount, resultError
}

func UpscaleDaemonsets(ctx context.Context, clientset kubernetes.Interface, daemonsets *v1.DaemonSetList) (int, error) {
	upscaledCount := 0
	var resultError error
	for _, d := range daemonsets.Items {
		if _, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]; exists {
			log.Infof("Scaling up daemonset %s", d.Name)
			delete(d.Spec.Template.Spec.NodeSelector, noscheduleAnnotation)
			_, err := clientset.AppsV1().DaemonSets(d.Namespace).Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling up resource %s: %w", d.GetName(), err), resultError)
			} else {
				upscaledCount++
			}
		} else {
			log.Infof("Daemonset %s is not marked as downscaled", d.Name)
		}
	}
	return upscaledCount, resultError
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
