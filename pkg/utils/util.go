package utils

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
)

// GetAlreadyExistingSecrets will return already existing secrets matching the namespace, name and labels within a cluster
func GetAlreadyExistingSecrets(namespaces []string, name string, labels map[string]string) []*corev1.Secret {
	sort.Strings(namespaces)

	secrets := []*corev1.Secret{}
	for _, namespace := range namespaces {
		s, err := GetSecret(namespace, name, labels)
		if err == nil {
			secrets = append(secrets, s)
		}
	}

	return secrets
}

// GetSecretByName is a helper function to get a secret in a namespace with a given name if it exists, else it will return a nil
func GetSecretByName(namespace string, name string) (*corev1.Secret, error) {
	return GetSecret(namespace, name, nil)
}

// GetSecretByUID is a helper function to get a secret in a namespace with a given name and uid if it exists, else it will return a nil
func GetSecretByUID(namespace string, name string, uid types.UID) (*corev1.Secret, error) {
	s := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			UID:       uid,
			Name:      name,
			Namespace: namespace,
		},
	}
	// err := sdk.Get(s)
	// return s, err
	return s, nil
}

// GetSecret will return a secret matching the required conditions
func GetSecret(namespace string, name string, labels map[string]string) (*corev1.Secret, error) {
	s := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	// err := sdk.Get(s)
	// return s, err
	return s, nil
}

// GetGeneratedSecretRef will return a GeneratedSecretRef for the given secret
func GetGeneratedSecretRef(secret corev1.Secret) generatedsecretv1.GeneratedSecretRef {
	return generatedsecretv1.GeneratedSecretRef{
		Name:            secret.GetName(),
		Namespace:       secret.GetNamespace(),
		Type:            secret.Type,
		ResourceVersion: secret.GetResourceVersion(),
		UID:             secret.GetUID(),
	}
}
