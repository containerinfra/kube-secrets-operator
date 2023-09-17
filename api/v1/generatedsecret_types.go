package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type SecretType string

const (
	SecretTypeOpaque    SecretType = "Opaque"
	SecretTypeBinary    SecretType = "binary"
	SecretTypeBasicAuth SecretType = "basic-auth"
	SecretTypeSSHAuth   SecretType = "ssh-auth"
)

// GeneratedSecretSpec defines the desired state of Secret
type GeneratedSecretSpec struct {
	// SecretType holds the type of secret being generated
	SecretType SecretType `json:"secretType"`

	// Metadata holds the metadata for the kubernetes secret generation
	Metadata SecretMetadata `json:"metadata"`

	// Template is used for composing the secret. Only necessary when using SecretType Opague (default)
	// +optional
	Template SecretTemplate `json:"template"`

	// // SecretRef is an reference to a kubernetes secret that will be created
	// SecretRef *corev1.SecretReference `json:"passwordSecretRef,omitempty"`
}

type SecretMetadata struct {
	// Name is the name of the Kubernetes secret being created
	Name string `json:"name"`
	// Namespaces is a list of namesapces in which the secret will be generated
	// +optional
	Namespaces []string `json:"namespaces"`

	// +optional
	Annotations map[string]string `json:"annotations"`

	// +optional
	Labels map[string]string `json:"labels"`

	// +optional
	Type corev1.SecretType `json:"type"`
}

// GetName returns the name of the generated secret
func (meta *SecretMetadata) GetName() string {
	return meta.Name
}

// GetNamespaces returns a list of namespaces for which this secret will be generated
func (meta *SecretMetadata) GetNamespaces() []string {
	return meta.Namespaces
}

// GetLabels returns a hashmap of the labels to be added to the generated secret
func (meta *SecretMetadata) GetLabels() map[string]string {
	return meta.Labels
}

// GetAnnotations returns a hashmap of annotations to be added to the generated secret
func (meta *SecretMetadata) GetAnnotations() map[string]string {
	return meta.Annotations
}

type SecretTemplate struct {
	Data SecretValueItems `json:"data"`
}

type SecretValueItems []SecretValueItemTemplate

type SecretValueItemTemplate struct {
	Name string `json:"name"`
	// +optional
	Value string `json:"value"`
	// +optional
	Length uint32 `json:"length,omitempty"`

	// +optional
	MinLength uint32 `json:"minLength,omitempty"`
	// +optional
	MaxLength uint32 `json:"maxLength,omitempty"`

	// +optional
	MaxSymbols uint32 `json:"maxSymbols"`

	// +optional
	MaxDigits uint32 `json:"maxDigits"`

	// +optional
	NoUpper bool `json:"noUpperCaseValues"`

	// +optional
	NoRepeat bool `json:"noRepeatedValues"`
}

// GeneratedSecretStatus defines the observed state of Secret.
type GeneratedSecretStatus struct {
	Initalized          bool                `json:"initalized"`
	Status              string              `json:"status"`
	SecretsGeneratedRef GeneratedSecretsRef `json:"secretsGeneratedRef"`
}

// GeneratedSecretsRef is a list of references to secrets
type GeneratedSecretsRef struct {
	Secrets []GeneratedSecretRef `json:"secrets"`
}

// GeneratedSecretRef describes a reference to a generated secret by referencing meta data
type GeneratedSecretRef struct {
	Name            string            `json:"name"`
	Type            corev1.SecretType `json:"type"`
	Namespace       string            `json:"namespace"`
	ResourceVersion string            `json:"resourceVersion"`
	UID             types.UID         `json:"uid"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GeneratedSecret is the Schema for the Secrets API
type GeneratedSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GeneratedSecretSpec   `json:"spec,omitempty"`
	Status GeneratedSecretStatus `json:"status,omitempty"`
}

// GetSecretName returns the name of the generated secret or if non provided, fallsback to the the name of the password resource
func (s *GeneratedSecret) GetSecretName() string {
	if s.Spec.Metadata.GetName() != "" {
		return s.Spec.Metadata.GetName()
	}
	return s.GetName()
}

// GetSecretLabels returns a hashmap of the labels to be added to the generated secret
func (s *GeneratedSecret) GetSecretLabels() map[string]string {
	return s.Spec.Metadata.GetLabels()
}

// GetSecretAnnotations returns a hashmap of annotations to be added to the generated secret
func (s *GeneratedSecret) GetSecretAnnotations() map[string]string {
	return s.Spec.Metadata.GetAnnotations()
}

//+kubebuilder:object:root=true

// GeneratedSecretList contains a list of GeneratedSecrets
type GeneratedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GeneratedSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GeneratedSecret{}, &GeneratedSecretList{})
}
