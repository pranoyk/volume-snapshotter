package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotAction struct {
	metav1.TypeMeta  `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotActionSpec  `json:"spec"`
}

type SnapshotActionSpec struct {
	Action         string `json:"action"`
	SnapshotName   string `json:"snapshotName"`
	SourcePVC      string `json:"sourcePVC"`
	DestinationPVC string `json:"destinationPVC"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SnapshotActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SnapshotAction `json:"items"`
}