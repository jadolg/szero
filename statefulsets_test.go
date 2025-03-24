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

func TestGetStatefulsets(t *testing.T) {
	ctx := context.Background()
	clientset := testclient.NewClientset()
	statefulsets, err := getStatefulSets(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, statefulsets.Items, 0)

	statefulset := v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	_, err = clientset.AppsV1().StatefulSets("default").Create(ctx, &statefulset, metav1.CreateOptions{})
	assert.NoError(t, err)

	newStatefulsets, err := getStatefulSets(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, newStatefulsets.Items, 1)
}

func TestDownscaleStatefulSets(t *testing.T) {
	testCases := []struct {
		name               string
		statefulset        v1.StatefulSet
		expectedDownscaled int
		expectedReplicas   int32
		expectedOldScale   string
	}{
		{
			name: "When the statefulset was not previously downscaled then it is downscaled",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: v1.StatefulSetSpec{
					Replicas: int32Ptr(2),
				},
			},
			expectedDownscaled: 1,
			expectedReplicas:   0,
			expectedOldScale:   "2",
		},
		{
			name: "When the statefulset was previously downscaled then nothing happens",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.StatefulSetSpec{
					Replicas: int32Ptr(0),
				},
			},
			expectedDownscaled: 0,
			expectedReplicas:   0,
			expectedOldScale:   "2",
		},
		{
			name: "When the statefulset has the downscaled annotation but the replicas are not 0 then the statefulset gets downscaled",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.StatefulSetSpec{
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

			_, err := clientset.AppsV1().StatefulSets("default").Create(ctx, &tc.statefulset, metav1.CreateOptions{})
			assert.NoError(t, err)

			statefulsets, err := getStatefulSets(ctx, clientset, "default")
			assert.NoError(t, err)

			downscaled, err := downscaleStatefulSets(ctx, clientset, statefulsets)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDownscaled, downscaled)

			newDeployments, err := getStatefulSets(ctx, clientset, "default")
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

func TestUpscaleStatefulSets(t *testing.T) {
	testCases := []struct {
		name             string
		statefulset      v1.StatefulSet
		expectedUpscaled int
		expectedReplicas int32
	}{
		{
			name: "When the service was previously downscaled then it is upscaled",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Annotations: map[string]string{
						replicasAnnotation: "2",
					},
				},
				Spec: v1.StatefulSetSpec{
					Replicas: int32Ptr(0),
				},
			},
			expectedReplicas: 2,
			expectedUpscaled: 1,
		},
		{
			name: "When the service was previously was not downscaled then nothing happens",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.StatefulSetSpec{
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

			_, err := clientset.AppsV1().StatefulSets("default").Create(ctx, &tc.statefulset, metav1.CreateOptions{})
			assert.NoError(t, err)

			statefulsets, err := getStatefulSets(ctx, clientset, "default")
			assert.NoError(t, err)

			upscaled, err := upscaleStatefulSets(ctx, clientset, statefulsets)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUpscaled, upscaled)

			newDeployments, err := getStatefulSets(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDeployments.Items {
				assert.Equal(t, tc.expectedReplicas, *d.Spec.Replicas)
				_, present := d.Annotations[replicasAnnotation]
				assert.False(t, present)
			}
		})
	}
}

func TestRestartStatefulSets(t *testing.T) {
	testCases := []struct {
		name        string
		statefulset v1.StatefulSet
	}{
		{
			name: "When annotations are present then annotations are updated",
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.StatefulSetSpec{
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
			statefulset: v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.StatefulSetSpec{
					Replicas: int32Ptr(1),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().StatefulSets("default").Create(ctx, &tc.statefulset, metav1.CreateOptions{})
			assert.NoError(t, err)

			statefulsets, err := getStatefulSets(ctx, clientset, "default")
			assert.NoError(t, err)

			upscaled, err := restartStatefulSets(ctx, clientset, statefulsets)
			assert.NoError(t, err)
			assert.Equal(t, 1, upscaled)

			newDeployments, err := getStatefulSets(ctx, clientset, "default")
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
