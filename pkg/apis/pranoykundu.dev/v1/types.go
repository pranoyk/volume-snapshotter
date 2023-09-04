package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`
type SnapshotActions struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   SnapshotActionsSpec
	Status SnapshotActionsStatus
}

type SnapshotActionsSpec struct {
	Action         string
	SnapshotName   string
	SourcePVC      string
	DestinationPVC string
}

type SnapshotActionsStatus struct {
	Progress string
}