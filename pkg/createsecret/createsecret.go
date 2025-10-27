package createsecret

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretOptions struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	StringData  map[string]string
	Data        map[string][]byte
}

func ConstructSecret(opts SecretOptions) v1.Secret {
	return v1.Secret{
		Type: v1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name:        opts.Name,
			Namespace:   opts.Namespace,
			Labels:      opts.Labels,
			Annotations: opts.Annotations,
		},
		StringData: opts.StringData,
		Data:       opts.Data,
	}
}

func ExistSecret(ctx context.Context, client client.Client, namespace, name string) (bool, error) {
	if err := client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, &v1.Secret{}); err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func CreateOrUpdateSecret(ctx context.Context, cl client.Client, secret *v1.Secret) error {
	if err := cl.Create(ctx, secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "unable to create secret")
		}
		if err := cl.Update(ctx, secret); err != nil {
			return errors.Wrap(err, "unable to update secret")
		}
	}
	return nil
}
