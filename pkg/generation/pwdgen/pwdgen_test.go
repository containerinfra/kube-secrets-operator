package pwdgen

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/containerinfra/kube-secrets-operator/api/v1"
)

// mockSecretFetcher is a mock implementation of SecretFetcher for testing
type mockSecretFetcher struct {
	secrets map[string]*corev1.Secret
}

func (m *mockSecretFetcher) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	secret, ok := m.secrets[key.Namespace+"/"+key.Name]
	if !ok {
		return fmt.Errorf("secret not found: %s/%s", key.Namespace, key.Name)
	}

	// Type assert to *corev1.Secret and copy the data
	if s, ok := obj.(*corev1.Secret); ok {
		s.ObjectMeta = secret.ObjectMeta
		s.Data = secret.Data
		s.Type = secret.Type
		return nil
	}

	return fmt.Errorf("object is not a secret")
}

func newMockSecretFetcher() *mockSecretFetcher {
	return &mockSecretFetcher{
		secrets: make(map[string]*corev1.Secret),
	}
}

func (m *mockSecretFetcher) addSecret(namespace, name string, data map[string][]byte) {
	m.secrets[namespace+"/"+name] = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}
}

func TestGeneratePasswords(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: map[string]v1.SecretValueItemTemplate{
			"MY_PASSWORD": {
				Generated: &v1.GeneratedValueSpec{
					Length: 10,
				},
			},
		},
	})
	if data == nil {
		t.Errorf("generatePasswords did not return any output")
	}

	if len(data) != 1 {
		t.Errorf("Amount of items in data from generatePasswords was incorrect, got: %d, want: %d", len(data), 1)
	}

	password := string(data["MY_PASSWORD"])
	if len(password) != 10 {
		t.Errorf("Length of generated password was incorrect, got: %d, want: %d", len(password), 10)
	}
}

func TestGeneratePasswordsDifferent(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: map[string]v1.SecretValueItemTemplate{
			"MY_PASSWORD": {
				Generated: &v1.GeneratedValueSpec{
					Length: 10,
				},
			},
		},
	})

	data2 := GeneratePasswords(&v1.SecretTemplate{
		Data: map[string]v1.SecretValueItemTemplate{
			"MY_PASSWORD": {
				Generated: &v1.GeneratedValueSpec{
					Length: 10,
				},
			},
		},
	})
	password := string(data["MY_PASSWORD"])
	password2 := string(data2["MY_PASSWORD"])

	if password == password2 {
		t.Errorf("Password should be regenered, got: %s and : %s", password, password2)
	}

}

func TestGeneratePasswordsRandomLength(t *testing.T) {

	data := GeneratePasswords(&v1.SecretTemplate{
		Data: map[string]v1.SecretValueItemTemplate{
			"MY_PASSWORD": {
				Generated: &v1.GeneratedValueSpec{
					MinLength: 10,
					MaxLength: 32,
				},
			},
		},
	})
	password := string(data["MY_PASSWORD"])

	if len(password) < 10 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted a min of: %d", len(password), 10)
	}

	if len(password) > 32 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted a max of: %d", len(password), 32)
	}
}

func TestGeneratePasswordsRandomLengthDifferent(t *testing.T) {

	for count := 2; count < 100; count++ {
		data := GeneratePasswords(&v1.SecretTemplate{
			Data: map[string]v1.SecretValueItemTemplate{
				"MY_PASSWORD": {
					Generated: &v1.GeneratedValueSpec{
						MinLength: 10,
						MaxLength: uint32(count),
					},
				},
			},
		})
		password := string(data["MY_PASSWORD"])

		if count >= 10 {
			if len(password) < 10 {
				t.Errorf("Length of generated password was incorrect, got: %d, wanted a min of: %d", len(password), 10)
			}
		}

		if len(password) > count {
			t.Errorf("Length of generated password was incorrect, got: %d, wanted a max of: %d", len(password), count)
		}
	}

}

func TestGeneratePasswordsMaxLengthHasPreference(t *testing.T) {
	data := GeneratePasswords(&v1.SecretTemplate{
		Data: map[string]v1.SecretValueItemTemplate{
			"MY_PASSWORD": {
				Generated: &v1.GeneratedValueSpec{
					MinLength: 10,
					MaxLength: 2,
				},
			},
		},
	})
	password := string(data["MY_PASSWORD"])
	if len(password) != 2 {
		t.Errorf("Length of generated password was incorrect, got: %d, wanted: %d", len(password), 2)
	}
}

func TestGenerateValues(t *testing.T) {
	tests := []struct {
		name             string
		template         *v1.SecretTemplate
		setupMock        func(*mockSecretFetcher)
		defaultNamespace string
		expectedKeys     []string
		expectError      bool
		validate         func(*testing.T, map[string][]byte)
	}{
		{
			name: "direct value field",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"api-key": {
						Value: "my-api-key",
					},
				},
			},
			setupMock:        func(m *mockSecretFetcher) {},
			defaultNamespace: "default",
			expectedKeys:     []string{"api-key"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, "my-api-key", string(data["api-key"]))
			},
		},
		{
			name: "static value (legacy)",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"static-key": {
						Static: &v1.StaticValueSpec{
							Value: "static-value",
						},
					},
				},
			},
			setupMock:        func(m *mockSecretFetcher) {},
			defaultNamespace: "default",
			expectedKeys:     []string{"static-key"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, "static-value", string(data["static-key"]))
			},
		},
		{
			name: "generated value",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"generated-key": {
						Generated: &v1.GeneratedValueSpec{
							Length: 16,
						},
					},
				},
			},
			setupMock:        func(m *mockSecretFetcher) {},
			defaultNamespace: "default",
			expectedKeys:     []string{"generated-key"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, 16, len(data["generated-key"]))
			},
		},
		{
			name: "templated value with input secret",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"connection-string": {
						Templated: &v1.TemplatedValueSpec{
							Template: "postgres://{{.Ref.username}}:{{.Ref.password}}@{{.Ref.host}}:{{.Ref.port}}/{{.Ref.database}}",
							InputSecretRef: &v1.SecretReference{
								Name:      "db-credentials",
								Namespace: "default",
							},
						},
					},
				},
			},
			setupMock: func(m *mockSecretFetcher) {
				m.addSecret("default", "db-credentials", map[string][]byte{
					"username": []byte("dbuser"),
					"password": []byte("dbpass"),
					"host":     []byte("localhost"),
					"port":     []byte("5432"),
					"database": []byte("mydb"),
				})
			},
			defaultNamespace: "default",
			expectedKeys:     []string{"connection-string"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, "postgres://dbuser:dbpass@localhost:5432/mydb", string(data["connection-string"]))
			},
		},
		{
			name: "templated value with default namespace",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"config": {
						Templated: &v1.TemplatedValueSpec{
							Template: "API_KEY={{.Ref.key}}",
							InputSecretRef: &v1.SecretReference{
								Name: "api-secret",
							},
						},
					},
				},
			},
			setupMock: func(m *mockSecretFetcher) {
				m.addSecret("test-ns", "api-secret", map[string][]byte{
					"key": []byte("secret-key-123"),
				})
			},
			defaultNamespace: "test-ns",
			expectedKeys:     []string{"config"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, "API_KEY=secret-key-123", string(data["config"]))
			},
		},
		{
			name: "multiple values of different types",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"static": {
						Value: "static-value",
					},
					"generated": {
						Generated: &v1.GeneratedValueSpec{
							Length: 10,
						},
					},
					"templated": {
						Templated: &v1.TemplatedValueSpec{
							Template: "Value: {{.Ref.data}}",
							InputSecretRef: &v1.SecretReference{
								Name:      "input-secret",
								Namespace: "default",
							},
						},
					},
				},
			},
			setupMock: func(m *mockSecretFetcher) {
				m.addSecret("default", "input-secret", map[string][]byte{
					"data": []byte("templated-data"),
				})
			},
			defaultNamespace: "default",
			expectedKeys:     []string{"static", "generated", "templated"},
			expectError:      false,
			validate: func(t *testing.T, data map[string][]byte) {
				assert.Equal(t, "static-value", string(data["static"]))
				assert.Equal(t, 10, len(data["generated"]))
				assert.Equal(t, "Value: templated-data", string(data["templated"]))
			},
		},
		{
			name: "templated value with missing input secret",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"config": {
						Templated: &v1.TemplatedValueSpec{
							Template: "Value: {{.Ref.data}}",
							InputSecretRef: &v1.SecretReference{
								Name:      "missing-secret",
								Namespace: "default",
							},
						},
					},
				},
			},
			setupMock:        func(m *mockSecretFetcher) {},
			defaultNamespace: "default",
			expectError:      true,
		},
		{
			name: "templated value without input secret ref",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"config": {
						Templated: &v1.TemplatedValueSpec{
							Template: "Value: {{.Ref.data}}",
						},
					},
				},
			},
			setupMock:        func(m *mockSecretFetcher) {},
			defaultNamespace: "default",
			expectError:      true,
		},
		{
			name: "templated value with invalid template",
			template: &v1.SecretTemplate{
				Data: map[string]v1.SecretValueItemTemplate{
					"config": {
						Templated: &v1.TemplatedValueSpec{
							Template: "Value: {{.Ref.data",
							InputSecretRef: &v1.SecretReference{
								Name:      "input-secret",
								Namespace: "default",
							},
						},
					},
				},
			},
			setupMock: func(m *mockSecretFetcher) {
				m.addSecret("default", "input-secret", map[string][]byte{
					"data": []byte("value"),
				})
			},
			defaultNamespace: "default",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mock := newMockSecretFetcher()
			tt.setupMock(mock)

			data, err := GenerateValues(ctx, mock, tt.defaultNamespace, tt.template)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, data)

				// Check that all expected keys are present
				for _, key := range tt.expectedKeys {
					assert.Contains(t, data, key, "expected key %s to be in generated data", key)
				}

				// Run custom validation if provided
				if tt.validate != nil {
					tt.validate(t, data)
				}
			}
		})
	}
}
