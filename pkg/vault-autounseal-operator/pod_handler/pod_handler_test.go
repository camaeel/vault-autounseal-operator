package podhandler

import (
	"testing"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPodHandlerFunctions(t *testing.T) {
	cfg := config.Config{
		Namespace: "vault",
	}
	ret := GetPodHandlerFunctions(&cfg, nil)
	assert.NotNil(t, ret)
}

func TestIsInitialized_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-initialized": "true",
			},
		},
	}

	res := isInitialized(pod)
	assert.True(t, res)
}

func TestIsInitialized_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-initialized": "false",
			},
		},
	}

	res := isInitialized(pod)
	assert.False(t, res)
}

func TestIsInitialized_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isInitialized(pod)
	assert.False(t, res)
}

func TestIsInitialized_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-initialized": "invalid",
			},
		},
	}

	res := isInitialized(pod)
	assert.False(t, res)
}

func TestIsSealed_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-sealed": "true",
			},
		},
	}

	res := isSealed(pod)
	assert.True(t, res)
}

func TestIsSealed_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-sealed": "false",
			},
		},
	}

	res := isSealed(pod)
	assert.False(t, res)
}

func TestIsSealed_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isSealed(pod)
	assert.False(t, res)
}

func TestIsSealed_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-sealed": "invalid",
			},
		},
	}

	res := isSealed(pod)
	assert.False(t, res)
}

func TestIsLeader_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-active": "true",
			},
		},
	}

	res := isLeader(pod)
	assert.True(t, res)
}

func TestIsLeader_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-active": "false",
			},
		},
	}

	res := isLeader(pod)
	assert.False(t, res)
}

func TestIsLeader_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isLeader(pod)
	assert.False(t, res)
}

func TestIsLeader_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-active": "invalid",
			},
		},
	}

	res := isLeader(pod)
	assert.False(t, res)
}
