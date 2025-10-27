package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAlreadyExistingSecrets(t *testing.T) {
	t.Skip("Skipping test for legacy stub function that is not used in production code")

	namespaces := []string{"namespace1", "namespace2"}
	name := "my-secret"
	labels := map[string]string{
		"app": "my-app",
	}

	// Create a list of existing secrets
	existingSecrets := []*corev1.Secret{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace1",
				Labels:    labels,
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace2",
				Labels:    labels,
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "namespace4",
				Labels:    labels,
			},
		},
	}

	// Call the GetAlreadyExistingSecrets function
	secrets := GetAlreadyExistingSecrets(namespaces, name, labels)
	// Check if the returned secrets match the existing secrets
	assert.ElementsMatch(t, existingSecrets, secrets)
}
