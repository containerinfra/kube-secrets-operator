package generatedsecret

import (
	"context"
	"fmt"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	"github.com/containerinfra/kube-secrets-operator/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *GeneratedSecretReconciler) validateSpec(o generatedsecretv1.GeneratedSecret) error {
	if len(o.Spec.Metadata.Namespaces) == 0 {
		// recorder.Event(o, corev1.EventTypeWarning, "Validation failed", "Missing namespaces. Must be > 0")
		return fmt.Errorf("missing namespaces")
	}
	return nil
}

// getExpectedSecretKeys returns the list of keys that should exist in the secret data
func getExpectedSecretKeys(generatedSecret generatedsecretv1.GeneratedSecret) []string {
	keys := []string{}
	for key := range generatedSecret.Spec.Template.Data {
		keys = append(keys, key)
	}
	return keys
}

// ReconcileToSpec manages the lifecycle of secrets after creation and before deletion
// It will patch secrets with new metadata, and if necessary, will rotate the secrets to new values
func (r *GeneratedSecretReconciler) reconcileToSpec(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret) bool {
	logger := log.FromContext(ctx)

	updated := false
	generatedSecretsRefs := []generatedsecretv1.GeneratedSecretRef{}

	for _, secretRef := range generatedSecret.Status.SecretsGeneratedRef.Secrets {
		secret := &corev1.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{
			Name:      secretRef.Name,
			Namespace: secretRef.Namespace,
		}, secret)

		if err != nil {
			if errors.IsNotFound(err) {
				logger.Info(fmt.Sprintf("A managed resource is deleted: %s/%s @ %s", secretRef.Namespace, secretRef.Name, secretRef.UID))
				continue
			}

			logger.Info(fmt.Sprintf("failed to fetch secret reference: %s", err.Error()))
			generatedSecretsRefs = append(generatedSecretsRefs, secretRef)
			continue
		}

		// Check if secret data is missing or empty
		if len(secret.Data) == 0 {
			logger.Info("secret data is empty, marking for regeneration", "secret", secret.GetName(), "namespace", secret.GetNamespace())
			// Remove this secret from refs so it will be recreated in createMissingPasswordSecrets
			continue
		}

		// Verify that the secret has all expected data keys from the template
		expectedKeys := getExpectedSecretKeys(generatedSecret)
		missingKeys := []string{}
		for _, key := range expectedKeys {
			if _, exists := secret.Data[key]; !exists {
				missingKeys = append(missingKeys, key)
			}
		}

		if len(missingKeys) > 0 {
			logger.Info("secret is missing expected data keys, marking for regeneration", "secret", secret.GetName(), "namespace", secret.GetNamespace(), "missingKeys", missingKeys)
			// Remove this secret from refs so it will be recreated
			continue
		}

		// Prepare expected labels and annotations with ownership labels
		expectedLabels := generatedSecret.GetSecretLabels()
		if expectedLabels == nil {
			expectedLabels = make(map[string]string)
		}
		for k, v := range getLabelsForSecret(generatedSecret) {
			expectedLabels[k] = v
		}

		expectedAnnotations := generatedSecret.GetSecretAnnotations()
		if expectedAnnotations == nil {
			expectedAnnotations = make(map[string]string)
		}

		labelsEqual := equality.Semantic.DeepEqual(secret.GetLabels(), expectedLabels)
		annotationsEqual := equality.Semantic.DeepEqual(secret.GetAnnotations(), expectedAnnotations)

		if !labelsEqual || !annotationsEqual {
			logger.Info("secret labels or annotations do not match. Updating secret", "secret", secret.GetName(), "labelsEqual", labelsEqual, "annotationsEqual", annotationsEqual)
			secret.SetLabels(expectedLabels)
			secret.SetAnnotations(expectedAnnotations)
			err := r.Client.Update(ctx, secret)
			if err != nil {
				generatedSecretsRefs = append(generatedSecretsRefs, secretRef)
				logger.Info(fmt.Sprintf("Failed to reconcile a secret due to k8s api error: %s", err.Error()))
				continue
			}

			// Fetch the updated secret to get the latest UID and ResourceVersion
			updatedSecret := &corev1.Secret{}
			err = r.Client.Get(ctx, types.NamespacedName{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			}, updatedSecret)
			if err != nil {
				logger.Error(err, "failed to fetch updated secret", "secret", secret.GetName())
				// Use the existing secretRef if we can't fetch the updated one
				generatedSecretsRefs = append(generatedSecretsRefs, secretRef)
				continue
			}

			newSecretRef := utils.GetGeneratedSecretRef(*updatedSecret)
			generatedSecretsRefs = append(generatedSecretsRefs, newSecretRef)

			updated = true
		} else {
			// Even when not updating, use the current secret's metadata to ensure UID/ResourceVersion are up to date
			currentSecretRef := utils.GetGeneratedSecretRef(*secret)
			generatedSecretsRefs = append(generatedSecretsRefs, currentSecretRef)
		}
	}

	if !equality.Semantic.DeepEqual(generatedSecret.Status.SecretsGeneratedRef.Secrets, generatedSecretsRefs) {
		generatedSecret.Status.SecretsGeneratedRef.Secrets = generatedSecretsRefs
		err := r.updateStatusOrRetry(ctx, &generatedSecret)
		if err != nil {
			logger.Error(err, "failed to reconcile generated secret")
			return false
		}
	}

	return updated
}

// FetchExistingSecrets returns a list of existing password secrets in the cluster
func (r *GeneratedSecretReconciler) fetchExistingSecrets(ctx context.Context, o generatedsecretv1.GeneratedSecret) ([]corev1.Secret, []corev1.Secret) {
	logger := log.FromContext(ctx)

	invalidSecrets := []corev1.Secret{}
	validSecrets := []corev1.Secret{}

	for _, secretRef := range o.Status.SecretsGeneratedRef.Secrets {
		secret := &corev1.Secret{}

		err := r.Client.Get(ctx, types.NamespacedName{
			Name:      secretRef.Name,
			Namespace: secretRef.Namespace,
		}, secret)

		if err != nil {
			// TODO: handle more errors
			if errors.IsNotFound(err) {
				logger.Info(fmt.Sprintf("a managed resource is deleted: %s/%s @ %s", secretRef.Namespace, secretRef.Name, secretRef.UID))
				continue
			}
			logger.Error(err, "Secret get errored")
			// validSecrets = append(validSecrets, *secret)
			continue
		}

		if secret.UID != secretRef.UID {
			logger.Info(fmt.Sprintf("Secret UID invalid. Found: %s, expected: %s", secret.UID, secretRef.UID))
			// TODO Make as invalid reference, possibly resync?
			invalidSecrets = append(invalidSecrets, *secret)
			continue
		}
		if secret.GetResourceVersion() != secretRef.ResourceVersion {
			logger.Info(fmt.Sprintf("Secret ResourceVersion invalid. Found: %s, expected: %s", secret.GetResourceVersion(), secretRef.ResourceVersion))
			// TODO Make as invalid reference, possibly resync?
			// Someone probably manually edited, or this resource got modified by an external system
			invalidSecrets = append(invalidSecrets, *secret)
			continue
		}

		if secret.Type != corev1.SecretType(secretRef.Type) {
			logger.Info(fmt.Sprintf("Secret Type invalid. Found: %s, expected: %s", secret.Type, secretRef.Type))
			// Should never happen, this means we got a bug in our code or someone manually edited various resources; this should however update the resource version
			invalidSecrets = append(invalidSecrets, *secret)
			continue
		}
		validSecrets = append(validSecrets, *secret)
	}
	return validSecrets, invalidSecrets
}
