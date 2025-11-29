package pkg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

func UpscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	var resultError error
	upscaledCount := 0
	for _, d := range deployments.Items {
		upscaled, err := upscaleDeployment(ctx, clientset, d.Namespace, d.Name)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling up deployment %s: %v", d.Name, err), resultError)
		}
		if upscaled {
			upscaledCount++
		}
	}
	return upscaledCount, resultError
}

func DownscaleDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList) (int, error) {
	var resultError error
	downscaledCount := 0
	for _, d := range deployments.Items {
		downscaled, err := downscaleDeployment(ctx, clientset, d.Namespace, d.Name)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error scaling down deployment %s: %v", d.Name, err), resultError)
		}
		if downscaled {
			downscaledCount++
		}
	}
	return downscaledCount, resultError
}

func GetDeployments(ctx context.Context, clientset kubernetes.Interface, namespace string) (*v1.DeploymentList, error) {
	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting deployments: %w", err)
	}
	return deployments, err
}

func IsDeploymentReady(ds *v1.Deployment, downscaled bool) bool {
	if downscaled {
		return ds.Status.Replicas == 0 && ds.Status.ReadyReplicas == 0
	}
	return ds.Status.AvailableReplicas == *ds.Spec.Replicas
}

func downscaleDeployment(ctx context.Context, clientset kubernetes.Interface, namespace string, name string) (bool, error) {
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		_, downscaled := d.Annotations[replicasAnnotation]
		if !downscaled || *d.Spec.Replicas > 0 {
			log.Infof("Scaling down deployment %s from %d replicas", d.Name, d.Status.Replicas)
			if !downscaled {
				d.Annotations[replicasAnnotation] = fmt.Sprintf("%d", *d.Spec.Replicas)
			}
			*d.Spec.Replicas = 0
			_, err := clientset.AppsV1().Deployments(d.Namespace).Update(ctx, d, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		} else {
			log.Infof("Deployment %s already downscaled", d.Name)
		}
		return nil
	})
	return w, err
}

func upscaleDeployment(ctx context.Context, clientset kubernetes.Interface, namespace string, name string) (bool, error) {
	w := false
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		d, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		replicas, downscaled := d.Annotations[replicasAnnotation]
		if downscaled {
			intReplicas, err := strconv.ParseInt(replicas, 10, 32)
			if err != nil {
				return fmt.Errorf("error converting replicas to int: %w", err)
			}
			log.Infof("Scaling up deployment %s to %d replicas", d.Name, intReplicas)
			*d.Spec.Replicas = int32(intReplicas)
			delete(d.Annotations, replicasAnnotation)
			_, err = clientset.AppsV1().Deployments(d.Namespace).Update(ctx, d, metav1.UpdateOptions{})
			if err == nil {
				w = true
			}
			return err
		} else {
			log.Infof("Deployment %s already scaled up", d.Name)
		}
		return nil
	})
	return w, err
}

func WaitForDeployments(ctx context.Context, clientset kubernetes.Interface, deployments *v1.DeploymentList, timeout time.Duration, downscaled bool) error {
	ticker := time.NewTicker(1 * time.Second)
	timeoutAfter := time.After(timeout)

	for {
		select {
		case <-timeoutAfter:
			return fmt.Errorf("timeout waiting for deployments to reconcile")
		case <-ticker.C:
			done := true
			for _, d := range deployments.Items {
				dp, err := clientset.AppsV1().Deployments(d.Namespace).Get(ctx, d.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("error getting deployment %s: %w", d.Name, err)
				}
				if IsDeploymentReady(dp, downscaled) {
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
