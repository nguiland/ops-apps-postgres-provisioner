package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PostgresInstance configures one PostgreSQL server managed by this operator.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=pginstance
// +kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.connection.host`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
type PostgresInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresInstanceSpec   `json:"spec"`
	Status PostgresInstanceStatus `json:"status,omitempty"`
}

type PostgresInstanceSpec struct {
	Connection     PostgresConnection `json:"connection"`
	AdminSecretRef SecretReference    `json:"adminSecretRef"`
	// Secret discovery is cluster-wide. A database Secret must carry both
	// managed=true and instance=<this resource's name>.
}

type PostgresConnection struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self.trim().size() > 0",message="must not be blank"
	Host string `json:"host"`

	// +kubebuilder:default=5432
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// +kubebuilder:validation:Enum=disable;allow;prefer;require;verify-ca;verify-full
	// +kubebuilder:default=prefer
	SSLMode string `json:"sslMode,omitempty"`
}

type SecretReference struct {
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self.trim().size() > 0",message="must not be blank"
	Name string `json:"name"`

	// Empty means the PostgresInstance namespace.
	Namespace string `json:"namespace,omitempty"`
}

type PostgresInstanceStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Databases          []DatabaseStatus   `json:"databases,omitempty"`
}

type DatabaseStatus struct {
	Namespace               string `json:"namespace"`
	SecretName              string `json:"secretName"`
	Database                string `json:"database,omitempty"`
	Username                string `json:"username,omitempty"`
	ObservedResourceVersion string `json:"observedResourceVersion,omitempty"`
	Ready                   bool   `json:"ready"`
	Message                 string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
type PostgresInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresInstance `json:"items"`
}
