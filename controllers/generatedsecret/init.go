package generatedsecret

import (
	"context"
	"fmt"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	"github.com/containerinfra/kube-secrets-operator/pkg/generation/pwdgen"
	"github.com/containerinfra/kube-secrets-operator/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func (r *GeneratedSecretReconciler) initalizeGeneratedSecret(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret) error {
	logger := log.FromContext(ctx)

	// Generate all secret values (static, generated, and templated)
	passwordData, err := pwdgen.GenerateValues(ctx, r.Client, generatedSecret.Namespace, &generatedSecret.Spec.Template)
	if err != nil {
		// Set error conditions
		changed := meta.SetStatusCondition(&generatedSecret.Status.Conditions, generatedSecret.NewCondition(generatedsecretv1.ConditionError, metav1.ConditionTrue, generatedsecretv1.ReasonGenerationFailed, fmt.Sprintf("Failed to generate secret values: %v", err)))
		changed = meta.SetStatusCondition(&generatedSecret.Status.Conditions, generatedSecret.NewCondition(generatedsecretv1.ConditionReady, metav1.ConditionFalse, generatedsecretv1.ReasonGenerationFailed, "Secret generation failed")) || changed
		if changed {
			if err := r.updateStatusOrRetry(ctx, &generatedSecret); err != nil {
				logger.Error(err, "Failed to update status with error condition")
			}
		}
		return fmt.Errorf("failed to generate secret values: %w", err)
	}

	// Create the k8s secrets
	secrets := generatePasswordSecrets(generatedSecret, passwordData)

	generatedSecretsRefs := []generatedsecretv1.GeneratedSecretRef{}
	hasErrors := false
	for i := range secrets {
		secret := &secrets[i] // Use pointer to avoid copying
		err := r.Client.Create(ctx, secret)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				logger.Info(fmt.Sprintf("Failed to create secret: %v", err))
				r.Recorder.Eventf(&generatedSecret, corev1.EventTypeWarning, "Failed secret create", "Error while attempting to create secret in namespace '%s': %s", secret.GetNamespace(), err.Error())
				hasErrors = true
				continue
			} else {
				logger.Info(fmt.Sprintf("A secret for %s in namespace %s already exists...", secret.GetName(), secret.GetNamespace()))

				// Fetch the existing secret and use that instead
				err := r.Client.Get(ctx, types.NamespacedName{
					Name:      secret.GetName(),
					Namespace: secret.GetNamespace(),
				}, secret)
				if err != nil {
					r.Recorder.Eventf(&generatedSecret, corev1.EventTypeWarning, "Failed secret create", "Error while attempting to link secret in namespace '%s': %s", secret.GetNamespace(), err.Error())
					hasErrors = true
					continue
				}
			}
		} else {
			logger.Info("created secret", "namespace", secret.Namespace, "name", secret.Name, "uid", secret.UID)
		}

		// Verify we have the required metadata
		if secret.UID == "" {
			logger.Error(fmt.Errorf("secret UID is empty after create/get"), "secret metadata incomplete", "name", secret.Name, "namespace", secret.Namespace)
			hasErrors = true
			continue
		}

		ref := utils.GetGeneratedSecretRef(*secret)

		r.Recorder.Eventf(&generatedSecret, corev1.EventTypeNormal, "Created secret", "Created a new secret in namespace '%s'", secret.GetNamespace())
		generatedSecretsRefs = append(generatedSecretsRefs, ref)
	}

	// Update the status
	generatedSecret.Status.Initalized = true
	generatedSecret.Status.SecretsGeneratedRef.Secrets = generatedSecretsRefs
	generatedSecret.Status.SecretsCount = len(generatedSecretsRefs)

	// Set conditions
	changed := false
	if hasErrors {
		changed = meta.SetStatusCondition(&generatedSecret.Status.Conditions, generatedSecret.NewCondition(generatedsecretv1.ConditionError, metav1.ConditionTrue, generatedsecretv1.ReasonGenerationFailed, "Some secrets failed to be created"))
		changed = meta.SetStatusCondition(&generatedSecret.Status.Conditions, generatedSecret.NewCondition(generatedsecretv1.ConditionReady, metav1.ConditionFalse, generatedsecretv1.ReasonGenerationFailed, fmt.Sprintf("Created %d of %d secrets", len(generatedSecretsRefs), len(secrets)))) || changed
	} else {
		// Clear error condition
		changed = meta.RemoveStatusCondition(&generatedSecret.Status.Conditions, generatedsecretv1.ConditionError)
		changed = meta.SetStatusCondition(&generatedSecret.Status.Conditions, generatedSecret.NewCondition(generatedsecretv1.ConditionReady, metav1.ConditionTrue, generatedsecretv1.ReasonSecretsGenerated, fmt.Sprintf("Successfully generated %d secret(s)", len(generatedSecretsRefs)))) || changed
	}
	if changed {
		if err := r.updateStatusOrRetry(ctx, &generatedSecret); err != nil {
			logger.Error(err, "Failed to update status")
			return err
		}
	}
	return nil
}
