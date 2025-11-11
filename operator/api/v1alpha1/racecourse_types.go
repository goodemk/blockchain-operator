/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required. Any new fields you add must have json tags for the fields to be serialized.

// The desired state of Racecourse instance
type RacecourseSpec struct {
	// Sets the container image
	// +optional
	Image ImageSpec `json:"image,omitempty"`

	// The number of racecourse pods to run
	// +kubebuilder:default=2
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Defines resource requests/limits on the pods
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// WalletService defines how to connect to the wallet service
	WalletService WalletServiceSpec `json:"walletService"`

	// The Ethereum address of the deployed Race contract
	// If empty, the application will deploy a new contract automatically
	// Must be either empty or valid 42-character hex string
	// +kubebuilder:validation:Pattern=`^(0x[a-fA-F0-9]{40})?$`
	// +optional
	ContractAddress string `json:"contractAddress,omitempty"`

	// The ingress configuration
	// +optional
	Ingress IngressSpec `json:"ingress,omitempty"`
}

// Defines the configuration for the container image
type ImageSpec struct {
	// The repository from which to pull the image
	// +kubebuilder:default="racecourse"
	// +optional
	Repository string `json:"repository,omitempty"`

	// Tags associated with the image
	// +kubebuilder:default="0.0.1"
	// +optional
	Tag string `json:"tag,omitempty"`

	// The image pull policy
	// +kubebuilder:default="IfNotPresent"
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	// +optional
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

// Defines how to connect to the wallet service
type WalletServiceSpec struct {
	// The name of the Kubernetes Service associated with the wallet service
	Name string `json:"name"`

	// The namespace where the wallet Service is located
	// If empty, it defaults to the same namespace as the Racecourse instance
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// The JSON-RPC port associated with the wallet service
	// +kubebuilder:default=8545
	// +optional
	Port int32 `json:"port,omitempty"`
}

// Defines the configuration for ingress
type IngressSpec struct {
	// Determines if an ingress resource should be created
	// +kubebuilder:default=true
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// The name of the ingress class
	// +kubebuilder:default="nginx"
	// +optional
	ClassName string `json:"className,omitempty"`

	// The hostname for the ingress
	// +optional
	Host string `json:"host,omitempty"`

	// The path for the ingress
	// +optional
	Path string `json:"path,omitempty"`

	// Additional pod-level annotations for the ingress resource
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// The observed state of Racecourse
type RacecourseStatus struct {
	// The current phase of the Racecourse instance
	// +optional
	Phase RacecoursePhase `json:"phase,omitempty"`

	// The latest available observations of the Racecourse instance's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Indicates whether the Deployment is ready
	// +optional
	DeploymentReady bool `json:"deploymentReady,omitempty"`

	// The number of ready pods
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// The URL where the racecourse app can be accessed
	// +optional
	URL string `json:"url,omitempty"`

	// The resolved wallet service endpoint URL
	// +optional
	WalletServiceEndpoint string `json:"walletServiceEndpoint,omitempty"`
}

// The statusphase of a Racecourse instance
// +kubebuilder:validation:Enum=Pending;Running;Failed;Unknown
type RacecoursePhase string

const (
	// The Racecourse instance is being created
	RacecoursePhasePending RacecoursePhase = "Pending"
	// The Racecourse instance is running successfully
	RacecoursePhaseRunning RacecoursePhase = "Running"
	// The Racecourse instance has failed
	RacecoursePhaseFailed RacecoursePhase = "Failed"
	// The state of the Racecourse instance is unknown
	RacecoursePhaseUnknown RacecoursePhase = "Unknown"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rcs
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.status.url`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// The actual custom resource that users will create
type Racecourse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RacecourseSpec   `json:"spec,omitempty"`
	Status RacecourseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// A list of Racecourse instances
type RacecourseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Racecourse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Racecourse{}, &RacecourseList{})
}
