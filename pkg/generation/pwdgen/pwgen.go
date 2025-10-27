package pwdgen

import (
	"context"
	"fmt"
	"math"
	"math/rand"

	password "github.com/sethvargo/go-password/password"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/containerinfra/kube-secrets-operator/api/v1"
	"github.com/containerinfra/kube-secrets-operator/pkg/generation/templated"
)

// SecretFetcher is an interface for fetching secrets from Kubernetes
type SecretFetcher interface {
	Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error
}

// GeneratePasswords generates a series of random string based on the supplied password templates
// This function only handles static and generated values (not templated)
func GeneratePasswords(passwordSpec *v1.SecretTemplate) map[string][]byte {
	data := map[string][]byte{}

	for name, item := range passwordSpec.Data {
		// Direct value field (preferred)
		if item.Value != "" {
			data[name] = []byte(item.Value)
			continue
		}
		// Legacy static value (for backward compatibility)
		if item.Static != nil && item.Static.Value != "" {
			data[name] = []byte(item.Static.Value)
			continue
		}
		if item.Generated == nil {
			continue
		}
		passwordLength := getPasswordLength(&item)
		generatedPassword, err := password.Generate(passwordLength, getNumberOfDigits(&item), getNumberOfSymbols(&item), item.Generated.NoUpper, !item.Generated.NoRepeat)
		if err != nil {
			panic(err)
		}
		data[name] = []byte(generatedPassword)
	}
	return data
}

// GenerateValues generates all secret values including templated ones
// This requires access to the Kubernetes client to fetch input secrets for templating
func GenerateValues(ctx context.Context, fetcher SecretFetcher, defaultNamespace string, passwordSpec *v1.SecretTemplate) (map[string][]byte, error) {
	data := map[string][]byte{}

	for name, item := range passwordSpec.Data {
		// Handle direct value field (preferred)
		if item.Value != "" {
			data[name] = []byte(item.Value)
			continue
		}

		// Handle legacy static values (for backward compatibility)
		if item.Static != nil && item.Static.Value != "" {
			data[name] = []byte(item.Static.Value)
			continue
		}

		// Handle templated values
		if item.Templated != nil {
			value, err := generateTemplatedValue(ctx, fetcher, defaultNamespace, item.Templated)
			if err != nil {
				return nil, fmt.Errorf("failed to generate templated value for key %s: %w", name, err)
			}
			data[name] = value
			continue
		}

		// Handle generated values
		if item.Generated != nil {
			passwordLength := getPasswordLength(&item)
			generatedPassword, err := password.Generate(passwordLength, getNumberOfDigits(&item), getNumberOfSymbols(&item), item.Generated.NoUpper, !item.Generated.NoRepeat)
			if err != nil {
				return nil, fmt.Errorf("failed to generate password for key %s: %w", name, err)
			}
			data[name] = []byte(generatedPassword)
			continue
		}
	}

	return data, nil
}

// generateTemplatedValue fetches the input secret and renders the template
func generateTemplatedValue(ctx context.Context, fetcher SecretFetcher, defaultNamespace string, spec *v1.TemplatedValueSpec) ([]byte, error) {
	if spec.InputSecretRef == nil {
		return nil, fmt.Errorf("inputSecretRef is required for templated values")
	}

	// Determine the namespace to fetch from
	namespace := spec.InputSecretRef.Namespace
	if namespace == "" {
		namespace = defaultNamespace
	}

	// Fetch the input secret
	var secret corev1.Secret
	err := fetcher.Get(ctx, types.NamespacedName{
		Name:      spec.InputSecretRef.Name,
		Namespace: namespace,
	}, &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch input secret %s/%s: %w", namespace, spec.InputSecretRef.Name, err)
	}

	// Render the template
	result, err := templated.RenderTemplate(spec.Template, secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return result, nil
}

func getPasswordLength(item *v1.SecretValueItemTemplate) int {
	if item.Generated == nil {
		return 0
	}
	lengthOfPassword := item.Generated.Length
	if item.Generated.MaxLength > 0 {
		lengthOfPassword = uint32(getRandomNumberBetween(int(item.Generated.MinLength), int(item.Generated.MaxLength)))
	} else {
		lengthOfPassword = uint32(math.Max(float64(lengthOfPassword), float64(item.Generated.MinLength)))
	}
	return int(lengthOfPassword)
}

func getNumberOfSymbols(item *v1.SecretValueItemTemplate) int {
	if item.Generated == nil {
		return 0
	}
	return int(math.Min(float64(getPasswordLength(item)), float64(getRandomNumberBetween(0, int(item.Generated.MaxSymbols)))))
}

func getNumberOfDigits(item *v1.SecretValueItemTemplate) int {
	return int(math.Min(float64(getPasswordLength(item)), float64(getRandomNumberBetween(0, int(item.Generated.MaxDigits)))))
}

func getRandomNumberBetween(min int, max int) int {
	if max == 0 {
		return 0
	} else if min >= max {
		return max
	}

	return min + int(rand.Int31n(int32(max-min)))
}
