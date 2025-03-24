package pkg

import (
	"context"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetNamespaces(ctx context.Context, clientset kubernetes.Interface) []string {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error(err)
		return []string{}
	}

	namespaceList := make([]string, len(namespaces.Items))
	for i, namespace := range namespaces.Items {
		namespaceList[i] = namespace.Name
	}
	return namespaceList
}
