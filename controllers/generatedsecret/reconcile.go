package generatedsecret

import (
	"context"
	"fmt"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *GeneratedSecretReconciler) reconcileExistingSecret(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret) error {
	logger := log.FromContext(ctx)

	// Figure out if any secret is in an invalid state and resync if necessary
	// An incorrect state is:
	//  - not the expected UID
	//  - different resource version
	//  - different resource type
	// If incorrect, see if we have any correct ones. If so, compare data and update any necessary values (i.g. the data field, labels, annotations, type)
	// if we cannot find any correct / valid secrets; mark this secret as invalid, report errors through k8s events -> support sending notifications through webhooks, maybe?
	validSecrets, invalidSecrets := r.fetchExistingSecrets(ctx, generatedSecret)

	// If there are no valid secrets, we should regenerate everything and go back to uninitalized
	if len(validSecrets) == 0 {
		logger.Info(fmt.Sprintf("No valid secrets found for: %s. Invalid count is: %d", generatedSecret.GetName(), len(invalidSecrets)), "result", "failed")
		// Generate everything again by setting the status to uninitalized
		generatedSecret.Status.Initalized = false
		logger.Info(fmt.Sprintf("Removing initalized status for: %s", generatedSecret.GetName()), "result", "failed")

		err := r.Client.Status().Update(ctx, &generatedSecret)
		if err != nil {
			return err
		}

		if len(invalidSecrets) > 0 {
			r.Recorder.Eventf(&generatedSecret, corev1.EventTypeWarning, "Missing secrets", "No valid secrets could be found, possibly due to invalid secrets (count is '%d')", len(invalidSecrets))
		} else {
			r.Recorder.Event(&generatedSecret, corev1.EventTypeWarning, "Missing secrets", "No valid secrets could be found")
		}
		return nil
	}

	if len(invalidSecrets) != 0 {
		for _, secret := range invalidSecrets {
			logger.Error(fmt.Errorf("secret has been externally modified: secret '%s' in namespace '%s'", secret.GetName(), secret.GetNamespace()), "error")
		}
	}

	// Update metadata if necessary
	if r.reconcileToSpec(ctx, generatedSecret) {
		return nil
	}
	return r.createMissingPasswordSecrets(ctx, generatedSecret, validSecrets)

}
