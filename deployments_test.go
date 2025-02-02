package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestGetDeployments(t *testing.T) {
	ctx := context.Background()
	clientset := testclient.NewClientset()
	deployments, err := getDeployments(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, deployments.Items, 0)

	deployment := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	_, err = clientset.AppsV1().Deployments("default").Create(ctx, &deployment, metav1.CreateOptions{})
	assert.NoError(t, err)

	newDeployments, err := getDeployments(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, newDeployments.Items, 1)
}

func TestDownscaleDeployments(t *testing.T) {
	testCases := []struct {
		name               string
		deployment         v1.Deployment
		expectedDownscaled int
		expectedReplicas   int32
		expectedOldScale   string
	}{
		{
			name: "When the deployment was not previously downscaled then it is downscaled",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(2),
				},
			},
			expectedDownscaled: 1,
			expectedReplicas:   0,
			expectedOldScale:   "2",
		},
		{
			name: "When the deployment was previously downscaled then nothing happens",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(0),
				},
			},
			expectedDownscaled: 0,
			expectedReplicas:   0,
			expectedOldScale:   "2",
		},
		{
			name: "When the deployment has the downscaled annotation but the replicas are not 0 then the deployment gets downscaled",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			expectedDownscaled: 1,
			expectedReplicas:   0,
			expectedOldScale:   "2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().Deployments("default").Create(ctx, &tc.deployment, metav1.CreateOptions{})
			assert.NoError(t, err)

			deployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			downscaled, err := downscaleDeployments(ctx, clientset, deployments)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDownscaled, downscaled)

			newDeployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDeployments.Items {
				assert.Equal(t, tc.expectedReplicas, *d.Spec.Replicas)
				oldScale, downscaled := d.Annotations[replicasAnnotation]
				assert.True(t, downscaled)
				assert.Equal(t, tc.expectedOldScale, oldScale)
			}
		})
	}
}

func TestUpscaleDeployments(t *testing.T) {
	testCases := []struct {
		name             string
		deployment       v1.Deployment
		expectedUpscaled int
		expectedReplicas int32
	}{
		{
			name: "When the service was previously downscaled then it is upscaled",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(0),
				},
			},
			expectedReplicas: 2,
			expectedUpscaled: 1,
		},
		{
			name: "When the service was previously was not downscaled then nothing happens",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
			expectedReplicas: 1,
			expectedUpscaled: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().Deployments("default").Create(ctx, &tc.deployment, metav1.CreateOptions{})
			assert.NoError(t, err)

			deployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			upscaled, err := upscaleDeployments(ctx, clientset, deployments)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUpscaled, upscaled)

			newDeployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDeployments.Items {
				assert.Equal(t, tc.expectedReplicas, *d.Spec.Replicas)
				_, present := d.Annotations[replicasAnnotation]
				assert.False(t, present)
			}
		})
	}
}

func int32Ptr(i int) *int32 {
	ptr := int32(i)
	return &ptr
}

func TestRestartDeployments(t *testing.T) {
	testCases := []struct {
		name       string
		deployment v1.Deployment
	}{
		{
			name: "When annotations are present then annotations are updated",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{},
						},
					},
				},
			},
		},
		{
			name: "When annotations are not present then annotations are initialized and then updated",
			deployment: v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.DeploymentSpec{
					Replicas: int32Ptr(1),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().Deployments("default").Create(ctx, &tc.deployment, metav1.CreateOptions{})
			assert.NoError(t, err)

			deployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			upscaled, err := restartDeployments(ctx, clientset, deployments)
			assert.NoError(t, err)
			assert.Equal(t, 1, upscaled)

			newDeployments, err := getDeployments(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDeployments.Items {
				cause, present := d.Spec.Template.Annotations[changeCauseAnnotation]
				assert.True(t, present)
				assert.Equal(t, "Restarted by szero", cause)
				restartedAt, present := d.Spec.Template.Annotations[restartedAtAnnotation]
				assert.True(t, present)
				assert.NotEmpty(t, restartedAt)
			}
		})
	}
}
