package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// -----------------------------------------------------------------------------
// Spec
// -----------------------------------------------------------------------------

// PodSvcSpec defines the desired state of PodSvc
type PodSvcSpec struct {
	// LabelName is the key of the label used to select Pods.
	// +kubebuilder:validation:Required
	LabelName string `json:"labelName"`

	// LabelValue is the value of the label used to select Pods.
	// +kubebuilder:validation:Required
	LabelValue string `json:"labelValue"`

	// Ports defines the desired service ports.
	// These are matched against Pod container ports.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Ports []corev1.ServicePort `json:"ports"`
}

// -----------------------------------------------------------------------------
// Status
// -----------------------------------------------------------------------------

// PodSvcStatusEntry describes the state of a single Pod managed by this CR.
type PodSvcStatusEntry struct {
	// PodName is the name of the Pod.
	PodName string `json:"podName"`

	// ServiceName is the name of the Service created for this Pod.
	ServiceName string `json:"serviceName"`

	// MatchedPorts are the ports that successfully matched Pod container ports.
	MatchedPorts []corev1.ServicePort `json:"matchedPorts,omitempty"`

	// UnmatchedPorts are CR ports that did not match any Pod container port.
	UnmatchedPorts []UnmatchedPortStatus `json:"unmatchedPorts,omitempty"`
}

// UnmatchedPortStatus describes a port from the CR that failed to match.
type UnmatchedPortStatus struct {
	Port     int32  `json:"port"`
	Name     string `json:"name,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// PodSvcStatus defines the observed state of PodSvc
type PodSvcStatus struct {
	// Entries contains one entry per Pod that matches the selector.
	Entries []PodSvcStatusEntry `json:"entries,omitempty"`
}

// -----------------------------------------------------------------------------
// Root types
// -----------------------------------------------------------------------------

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Label",type=string,JSONPath=".spec.labelName"
// +kubebuilder:printcolumn:name="Value",type=string,JSONPath=".spec.labelValue"
// +kubebuilder:printcolumn:name="Pods",type=integer,JSONPath=".status.entries[*].podName",description="Number of matching Pods"

// PodSvc is the Schema for the podsvcs API
type PodSvc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodSvcSpec   `json:"spec,omitempty"`
	Status PodSvcStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodSvcList contains a list of PodSvc
type PodSvcList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodSvc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodSvc{}, &PodSvcList{})
}
