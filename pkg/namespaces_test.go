package pkg

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetNamespaces(t *testing.T) {
	ctx := context.Background()
	clientset := testclient.NewClientset()
	namespaces := GetNamespaces(ctx, clientset)
	assert.Len(t, namespaces, 0)

	_, err := clientset.CoreV1().Namespaces().Create(
		ctx,
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
		},
		metav1.CreateOptions{},
	)
	assert.NoError(t, err)

	newNamespaces := GetNamespaces(ctx, clientset)
	assert.Len(t, newNamespaces, 1)
	assert.Equal(t, "test", newNamespaces[0])
}
