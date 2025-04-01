package pkg // Resource represents either a Deployment or StatefulSet
import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const replicasAnnotation = "szero/replicas"
const noscheduleAnnotation = "szero/noschedule"

func GetClientset(kubeconfig, context string) (*kubernetes.Clientset, error) {
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

func int32Ptr(i int) *int32 {
	ptr := int32(i)
	return &ptr
}
