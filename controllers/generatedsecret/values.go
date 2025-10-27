package generatedsecret

import (
	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	LabelGeneratedSecretName      = "generatedsecret.containerinfra.io/name"
	LabelGeneratedSecretNamespace = "generatedsecret.containerinfra.io/namespace"
	LabelGeneratedSecretRef       = "generatedsecret.containerinfra.io/ref"
)

func getLabelsForSecret(generatedSecret generatedsecretv1.GeneratedSecret) map[string]string {
	return map[string]string{
		LabelGeneratedSecretName:      generatedSecret.Name,
		LabelGeneratedSecretNamespace: generatedSecret.Namespace,
		LabelGeneratedSecretRef:       string(generatedSecret.UID),
	}
}

func isSecretOwnedBy(generatedSecret generatedsecretv1.GeneratedSecret, secret v1.Secret) bool {
	if secret.Labels == nil {
		return false
	}
	if secret.Labels[LabelGeneratedSecretName] != generatedSecret.Name {
		return false
	}
	if secret.Labels[LabelGeneratedSecretNamespace] != generatedSecret.Namespace {
		return false
	}
	if secret.Labels[LabelGeneratedSecretRef] != string(generatedSecret.UID) {
		return false
	}
	return true
}
