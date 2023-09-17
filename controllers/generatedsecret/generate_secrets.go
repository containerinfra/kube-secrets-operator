package generatedsecret

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	"github.com/containerinfra/kube-secrets-operator/pkg/createsecret"
	"github.com/containerinfra/kube-secrets-operator/pkg/utils"
)

// createMissingPasswordSecrets will create missing password secrets in the cluster
func (r *GeneratedSecretReconciler) createMissingPasswordSecrets(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret, validSecrets []corev1.Secret) error {
	logger := log.FromContext(ctx)

	if len(validSecrets) == 0 {
		return fmt.Errorf("invalid state: no valid secrets")
	}

	// copy the first secret over to all other namespaces
	validSecret := validSecrets[0]
	secretData := validSecret.Data

	// Now fetch the data and sync it to the new secrets
	secrets := generatePasswordSecrets(generatedSecret, secretData)

	initalLength := len(generatedSecret.Status.SecretsGeneratedRef.Secrets)

	// clear it, rebuild this list
	generatedSecret.Status.SecretsGeneratedRef.Secrets = []generatedsecretv1.GeneratedSecretRef{}

	for _, secret := range secrets {
		if secretIsListedInValidSecrets(validSecrets, secret) {
			generatedSecret.Status.SecretsGeneratedRef.Secrets = append(generatedSecret.Status.SecretsGeneratedRef.Secrets, utils.GetGeneratedSecretRef(secret))
			continue
		}

		err := r.Client.Create(ctx, &secret)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				logger.Info(fmt.Sprintf("Failed to create secret: %v, possibily a secret modified externally", err))
			}
			continue
		}
		r.Recorder.Eventf(&generatedSecret, corev1.EventTypeNormal, "Created secret", "Created a new secret in namespace '%s'", secret.GetNamespace())

		// New secret, so simply add it to the secrets ref
		generatedSecret.Status.SecretsGeneratedRef.Secrets = append(generatedSecret.Status.SecretsGeneratedRef.Secrets, utils.GetGeneratedSecretRef(secret))
	}

	if initalLength != len(generatedSecret.Status.SecretsGeneratedRef.Secrets) {
		return r.Client.Status().Update(ctx, &generatedSecret)
	}
	return nil
}

func secretIsListedInValidSecrets(validSecrets []corev1.Secret, secretCheck corev1.Secret) bool {
	for _, secret := range validSecrets {
		if secret.GetNamespace() == secretCheck.GetNamespace() && secret.GetName() == secretCheck.GetName() {
			return true
		}
	}
	return false
}

// generatePasswordSecrets
func generatePasswordSecrets(generatedSecret generatedsecretv1.GeneratedSecret, data map[string][]byte) []corev1.Secret {
	labels := generatedSecret.GetLabels()
	annotations := generatedSecret.GetSecretAnnotations()

	secrets := []corev1.Secret{}
	sort.Strings(generatedSecret.Spec.Metadata.GetNamespaces())

	secretType := corev1.SecretTypeOpaque
	if generatedSecret.Spec.Metadata.Type != "" {
		secretType = generatedSecret.Spec.Metadata.Type
	}

	for _, namespace := range generatedSecret.Spec.Metadata.GetNamespaces() {
		secret := createsecret.ConstructSecret(createsecret.SecretOptions{
			Name:        generatedSecret.GetSecretName(),
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
			Data:        data,
		})

		secret.Type = secretType
		secrets = append(secrets, secret)
	}
	return secrets
}
