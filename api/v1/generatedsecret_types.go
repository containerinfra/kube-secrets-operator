package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Condition types for GeneratedSecret
const (
	// ConditionReady indicates that the secret generation is ready and all secrets have been created
	ConditionReady = "Ready"
	// ConditionError indicates that there was an error during secret generation
	ConditionError = "Error"
)

// Condition reasons
const (
	ReasonSecretsGenerated    = "SecretsGenerated"
	ReasonGenerationFailed    = "GenerationFailed"
	ReasonTemplateError       = "TemplateError"
	ReasonInputSecretNotFound = "InputSecretNotFound"
	ReasonValidationFailed    = "ValidationFailed"
	ReasonReconciling         = "Reconciling"
)

type SecretType string

const (
	SecretTypeOpaque    SecretType = "Opaque"
	SecretTypeBinary    SecretType = "binary"
	SecretTypeBasicAuth SecretType = "basic-auth"
	SecretTypeSSHAuth   SecretType = "ssh-auth"
)

type DeletionPolicy string

const (
	DeleteOnCleanup DeletionPolicy = "Delete"
	RetainOnCleanup DeletionPolicy = "Retain"
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

	// DeletionPolicy is the policy to be used when the secret is deleted
	// +kubebuilder:default="Delete"
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy"`
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
	Labels map[string]string `json:"labels,omitempty"`

	// +optional
	Type string `json:"type,omitempty"`
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

type SecretValueItems map[string]SecretValueItemTemplate

type SecretValueItemTemplate struct {
	// Value is a static string value. This is a shorthand for static values.
	// +optional
	Value string `json:"value,omitempty"`

	// Static value is a static value that will be used to set the value of the secret
	// Deprecated: Use Value field instead for simpler syntax
	// +optional
	Static *StaticValueSpec `json:"static,omitempty"`

	// Templated value is a value that will be templated using the key-value pairs in the secret
	// +optional
	Templated *TemplatedValueSpec `json:"templated,omitempty"`

	// Generated value is a value that will be generated using a random string generator
	// +optional
	Generated *GeneratedValueSpec `json:"generated,omitempty"`
}

type SshKeyValueSpec struct {
	// Public Key is the public key that will be used to set the value of the secret
	PublicKey string `json:"publicKey"`
}

type StaticValueSpec struct {
	// Value is the static value that will be used to set the value of the secret
	Value string `json:"value"`
}

type TemplatedValueSpec struct {
	// Template is a string that will be templated using the key-value pairs in the secret. This is a go template string.
	Template string `json:"template"`
	// Input Secret reference is a reference to a secret that will be used to template the value
	// The value will be templated using the key-value pairs in the secret
	// +optional
	InputSecretRef *SecretReference `json:"inputSecretRef,omitempty"`
}

// SecretReference represents a reference to a Secret in a specific namespace
type SecretReference struct {
	// Name of the secret
	Name string `json:"name"`
	// Namespace of the secret. If empty, defaults to the namespace of the GeneratedSecret
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

type GeneratedValueSpec struct {
	// Lengt of the  generated value
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
	// Initalized indicates if the secret has been initialized
	// Deprecated: Use Conditions instead
	Initalized bool `json:"initalized"`

	// Status is a human-readable status message
	// Deprecated: Use Conditions instead
	Status string `json:"status"`

	// SecretsGeneratedRef contains references to all generated secrets
	SecretsGeneratedRef GeneratedSecretsRef `json:"secretsGeneratedRef"`

	// Conditions represent the latest available observations of the GeneratedSecret's state
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// SecretsCount is the total number of secrets that have been generated
	// +optional
	SecretsCount int `json:"secretsCount,omitempty"`
}

// GeneratedSecretsRef is a list of references to secrets
type GeneratedSecretsRef struct {
	Secrets []GeneratedSecretRef `json:"secrets"`
}

// GeneratedSecretRef describes a reference to a generated secret by referencing meta data
type GeneratedSecretRef struct {
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Namespace       string    `json:"namespace"`
	ResourceVersion string    `json:"resourceVersion"`
	UID             types.UID `json:"uid"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
//+kubebuilder:printcolumn:name="Secrets",type=integer,JSONPath=`.status.secretsCount`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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

// NewCondition creates a new condition for the GeneratedSecret
func (s *GeneratedSecret) NewCondition(conditionType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{
		Type:               conditionType,
		Status:             status,
		ObservedGeneration: s.Generation,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

func init() {
	SchemeBuilder.Register(&GeneratedSecret{}, &GeneratedSecretList{})
}
