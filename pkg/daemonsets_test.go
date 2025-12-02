package pkg

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestGetDaemonSets(t *testing.T) {
	ctx := context.Background()
	clientset := testclient.NewClientset()
	daemonsets, err := GetDaemonsets(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, daemonsets.Items, 0)

	daemonset := v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	_, err = clientset.AppsV1().DaemonSets("default").Create(ctx, &daemonset, metav1.CreateOptions{})
	assert.NoError(t, err)

	newDaemonSets, err := GetDaemonsets(ctx, clientset, "default")
	assert.NoError(t, err)
	assert.Len(t, newDaemonSets.Items, 1)
}

func TestDownscaleDaemonSets(t *testing.T) {
	testCases := []struct {
		name               string
		daemonset          v1.DaemonSet
		expectedDownscaled int
	}{
		{
			name: "When the daemonset was not previously downscaled then it is downscaled",
			daemonset: v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: v1.DaemonSetSpec{},
			},
			expectedDownscaled: 1,
		},
		{
			name: "When the daemonset was previously downscaled then nothing happens",
			daemonset: v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{
								noscheduleAnnotation: "true",
							},
						},
					},
				},
			},
			expectedDownscaled: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().DaemonSets("default").Create(ctx, &tc.daemonset, metav1.CreateOptions{})
			assert.NoError(t, err)

			daemonsets, err := GetDaemonsets(ctx, clientset, "default")
			assert.NoError(t, err)

			downscaled, err := DownscaleDaemonsets(ctx, clientset, daemonsets)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedDownscaled, downscaled)

			newDaemonsets, err := GetDaemonsets(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDaemonsets.Items {
				v, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]
				assert.True(t, exists)
				assert.Equal(t, "true", v)
			}
		})
	}
}

func TestUpscaleDaemonSets(t *testing.T) {
	testCases := []struct {
		name             string
		daemonset        v1.DaemonSet
		expectedUpscaled int
	}{
		{
			name: "When the daemonset was not previously downscaled then nothing happens",
			daemonset: v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: v1.DaemonSetSpec{},
			},
			expectedUpscaled: 0,
		},
		{
			name: "When the daemonset was previously downscaled then it is upscaled",
			daemonset: v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeSelector: map[string]string{
								noscheduleAnnotation: "true",
							},
						},
					},
				},
			},
			expectedUpscaled: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			clientset := testclient.NewClientset()

			_, err := clientset.AppsV1().DaemonSets("default").Create(ctx, &tc.daemonset, metav1.CreateOptions{})
			assert.NoError(t, err)

			daemonsets, err := GetDaemonsets(ctx, clientset, "default")
			assert.NoError(t, err)

			upscaled, err := UpscaleDaemonsets(ctx, clientset, daemonsets)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedUpscaled, upscaled)

			newDaemonsets, err := GetDaemonsets(ctx, clientset, "default")
			assert.NoError(t, err)

			for _, d := range newDaemonsets.Items {
				_, exists := d.Spec.Template.Spec.NodeSelector[noscheduleAnnotation]
				assert.False(t, exists)
			}
		})
	}
}
