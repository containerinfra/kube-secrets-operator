package generatedsecret

import (
	"context"
	"fmt"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *GeneratedSecretReconciler) reconcileGeneratedSecrets(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret) error {
	logger := log.FromContext(ctx)

	// Figure out if any secret is in an invalid state and resync if necessary
	// An incorrect state is:
	//  - not the expected UID
	//  - different resource version
	//  - different resource type
	// If incorrect, see if we have any correct ones. If so, compare data and update any necessary values (i.g. the data field, labels, annotations, type)
	// if we cannot find any correct / valid secrets; mark this secret as invalid, report errors through k8s events -> support sending notifications through webhooks, maybe?
	validSecrets, invalidSecrets := r.fetchExistingSecrets(ctx, generatedSecret)
	if len(invalidSecrets) != 0 {
		for _, secret := range invalidSecrets {
			logger.Error(fmt.Errorf("secret has been externally modified: secret '%s' in namespace '%s'", secret.GetName(), secret.GetNamespace()), "error")
		}

		secretRefs := []generatedsecretv1.GeneratedSecretRef{}
		// remove the invalid secrets from the status
		for _, secretRef := range generatedSecret.Status.SecretsGeneratedRef.Secrets {

			isValid := true
			for _, invalidSecret := range invalidSecrets {
				if secretRef.Namespace == invalidSecret.GetNamespace() && secretRef.Name == invalidSecret.GetName() {
					isValid = false // found in the invalid secrets list
					break
				}
			}
			if isValid {
				secretRefs = append(secretRefs, secretRef)
			}
		}

		if !equality.Semantic.DeepEqual(generatedSecret.Status.SecretsGeneratedRef.Secrets, secretRefs) {
			generatedSecret.Status.SecretsGeneratedRef.Secrets = secretRefs
			err := r.updateStatusOrRetry(ctx, &generatedSecret)
			if err != nil {
				return err
			}
		}
	}

	if len(validSecrets) == 0 && len(invalidSecrets) == 0 {
		return r.initalizeGeneratedSecret(ctx, generatedSecret)
	}

	// Update metadata if necessary
	if r.reconcileToSpec(ctx, generatedSecret) {
		return nil
	}
	return r.createMissingPasswordSecrets(ctx, generatedSecret, validSecrets)

}
