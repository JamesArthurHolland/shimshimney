package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ShimPodSpec defines the desired state of ShimPod
type ShimPodSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ShimPod. Edit shimpod_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ShimPodStatus defines the observed state of ShimPod
type ShimPodStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ShimPod is the Schema for the shimpods API
type ShimPod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShimPodSpec   `json:"spec,omitempty"`
	Status ShimPodStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ShimPodList contains a list of ShimPod
type ShimPodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShimPod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShimPod{}, &ShimPodList{})
}
