package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotActions struct {
	metav1.TypeMeta  `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotActionsSpec  `json:"spec"`
}

type SnapshotActionsSpec struct {
	Action         string `json:"action"`
	SnapshotName   string `json:"snapshotName"`
	SourcePVC      string `json:"sourcePVC"`
	DestinationPVC string `json:"destinationPVC"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotActionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SnapshotActions `json:"items"`
}