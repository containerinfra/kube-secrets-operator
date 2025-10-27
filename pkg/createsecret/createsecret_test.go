package createsecret

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestConstructSecret(t *testing.T) {
	opts := SecretOptions{
		Name:      "my-secret",
		Namespace: "my-namespace",
		Labels: map[string]string{
			"app": "my-app",
		},
		Annotations: map[string]string{
			"version": "v1",
		},
		StringData: map[string]string{
			"username": "admin",
			"password": "password123",
		},
		Data: map[string][]byte{
			"key": []byte("value"),
		},
	}

	secret := ConstructSecret(opts)

	assert.Equal(t, secret.Type, v1.SecretTypeOpaque)
	assert.Equal(t, secret.ObjectMeta.Name, opts.Name)
	assert.Equal(t, secret.ObjectMeta.Namespace, opts.Namespace)

	assert.Equal(t, secret.ObjectMeta.Labels, opts.Labels)
	assert.Equal(t, secret.ObjectMeta.Annotations, opts.Annotations)

	assert.Equal(t, secret.StringData, opts.StringData)
	assert.Equal(t, secret.Data, opts.Data)
}

func TestCreateOrUpdateSecret(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientBuilder().Build()
	namespace := uuid.NewString()
	name := uuid.NewString()

	t.Run("create", func(t *testing.T) {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		err := CreateOrUpdateSecret(ctx, client, secret)
		require.NoError(t, err)
		exists, err := ExistSecret(ctx, client, namespace, name)
		require.NoError(t, err)
		require.True(t, exists, "Expected secret to not exist after creation, but it exists")
	})

	t.Run("update", func(t *testing.T) {
		exists, err := ExistSecret(ctx, client, namespace, name)
		require.NoError(t, err)
		require.True(t, exists, "Expected secret to exist before update, but it does not exist")

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		secret.StringData = map[string]string{
			"username": "admin",
			"password": "newpassword",
		}
		err = CreateOrUpdateSecret(ctx, client, secret)
		require.NoError(t, err)
		// Verify secret update
		exists, err = ExistSecret(ctx, client, namespace, name)
		require.NoError(t, err)
		require.True(t, exists, "Expected secret to exist after update, but it does not exist")
	})
}

func TestExistSecret(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientBuilder().Build()
	namespace := uuid.NewString()
	name := uuid.NewString()

	t.Run("existing secret", func(t *testing.T) {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		err := client.Create(ctx, secret)
		require.NoError(t, err)

		exists, err := ExistSecret(ctx, client, namespace, name)
		require.NoError(t, err)
		require.True(t, exists, "Expected secret to exist, but it does not exist")
	})

	t.Run("non-existing secret", func(t *testing.T) {
		exists, err := ExistSecret(ctx, client, namespace, "does-not-exist")
		require.NoError(t, err)
		require.False(t, exists, "Expected secret to not exist, but it exists")
	})
}
