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

	"k8s.io/apimachinery/pkg/types"
)

func (r *GeneratedSecretReconciler) initalizeGeneratedSecret(ctx context.Context, generatedSecret generatedsecretv1.GeneratedSecret) error {
	logger := log.FromContext(ctx)

	passwordData := pwdgen.GeneratePasswords(&generatedSecret.Spec.Template)

	// Create the k8s secrets
	secrets := generatePasswordSecrets(generatedSecret, passwordData)

	generatedSecretsRefs := []generatedsecretv1.GeneratedSecretRef{}
	for _, secret := range secrets {
		err := r.Client.Create(ctx, &secret)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				logger.Info(fmt.Sprintf("Failed to create secret: %v", err))
				// TODO: handle error while attempting to create the secret
				r.Recorder.Eventf(&generatedSecret, corev1.EventTypeWarning, "Failed secret create", "Error while attempting to create secret in namespace '%s': %s", secret.GetNamespace(), err.Error())
				continue
			} else {
				logger.Info(fmt.Sprintf("A secret for %s in namespace %s already exists...", secret.GetName(), secret.GetNamespace()))

				// Fetch the existing secret and use that instead
				err := r.Client.Get(ctx, types.NamespacedName{
					Name:      secret.GetName(),
					Namespace: secret.GetNamespace(),
				}, &secret)
				if err != nil {
					r.Recorder.Eventf(&generatedSecret, corev1.EventTypeWarning, "Failed secret create", "Error while attempting to link secret in namespace '%s': %s", secret.GetNamespace(), err.Error())
					continue
				}
			}
		} else {
			fmt.Printf("created secret in ns %s/%s %q\n", secret.Namespace, secret.Name, secret.UID)
		}
		ref := utils.GetGeneratedSecretRef(secret)

		r.Recorder.Eventf(&generatedSecret, corev1.EventTypeNormal, "Created secret", "Created a new secret in namespace '%s'", secret.GetNamespace())
		generatedSecretsRefs = append(generatedSecretsRefs, ref)
	}

	// Update the status of the Password resource to reflect we initalized it
	generatedSecret.Status.Initalized = true
	generatedSecret.Status.SecretsGeneratedRef.Secrets = generatedSecretsRefs

	return r.Client.Status().Update(ctx, &generatedSecret)
}
