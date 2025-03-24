package main // Resource represents either a Deployment or StatefulSet
import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
	"time"
)

const replicasAnnotation = "szero/replicas"
const restartedAtAnnotation = "kubernetes.io/restartedAt"
const changeCauseAnnotation = "kubernetes.io/change-cause"

type Resource interface {
	GetName() string
	GetNamespace() string
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetReplicas() *int32
	SetReplicas(replicas *int32)
	GetTemplateAnnotations() map[string]string
	SetTemplateAnnotations(annotations map[string]string)
}

// ResourceUpdater handles the API update operation
type ResourceUpdater interface {
	Update(ctx context.Context, namespace string, resource Resource) error
}

func upscaleResource(ctx context.Context, resources []Resource, updater ResourceUpdater) (int, error) {
	var resultError error
	upscaledCount := 0

	for _, r := range resources {
		replicas, downscaled := r.GetAnnotations()[replicasAnnotation]
		if downscaled {
			intReplicas, err := strconv.ParseInt(replicas, 10, 32)
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error converting replicas to int: %w", err), resultError)
				continue
			}
			log.Infof("Scaling up resource %s to %d replicas", r.GetName(), intReplicas)
			r.SetReplicas(ptr(int32(intReplicas)))
			annotations := r.GetAnnotations()
			delete(annotations, replicasAnnotation)
			r.SetAnnotations(annotations)

			err = updater.Update(ctx, r.GetNamespace(), r)
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling up resource %s: %v", r.GetName(), err), resultError)
			} else {
				upscaledCount++
			}
		} else {
			log.Infof("Resource %s already scaled up", r.GetName())
		}
	}
	return upscaledCount, resultError
}

func downscaleResource(ctx context.Context, resources []Resource, updater ResourceUpdater) (int, error) {
	var resultError error
	downscaledCount := 0

	for _, r := range resources {
		_, downscaled := r.GetAnnotations()[replicasAnnotation]
		if !downscaled || *r.GetReplicas() > 0 {
			log.Infof("Scaling down resource %s", r.GetName())
			if !downscaled {
				annotations := r.GetAnnotations()
				if annotations == nil {
					annotations = make(map[string]string)
				}
				annotations[replicasAnnotation] = fmt.Sprintf("%d", *r.GetReplicas())
				r.SetAnnotations(annotations)
			}
			r.SetReplicas(ptr(int32(0)))

			err := updater.Update(ctx, r.GetNamespace(), r)
			if err != nil {
				resultError = errors.Join(fmt.Errorf("error scaling down resource %s: %v", r.GetName(), err), resultError)
			} else {
				downscaledCount++
			}
		} else {
			log.Infof("Resource %s already downscaled", r.GetName())
		}
	}
	return downscaledCount, resultError
}

func restartResource(ctx context.Context, resources []Resource, updater ResourceUpdater) (int, error) {
	var resultError error
	restartedCount := 0

	for _, r := range resources {
		log.Infof("Restarting resource %s", r.GetName())
		annotations := r.GetTemplateAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[restartedAtAnnotation] = time.Now().Format(time.RFC3339)
		annotations[changeCauseAnnotation] = "Restarted by szero"
		r.SetTemplateAnnotations(annotations)

		err := updater.Update(ctx, r.GetNamespace(), r)
		if err != nil {
			resultError = errors.Join(fmt.Errorf("error restarting resource %s: %v", r.GetName(), err), resultError)
		} else {
			restartedCount++
		}
	}
	return restartedCount, resultError
}

func ptr(i int32) *int32 {
	return &i
}

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
