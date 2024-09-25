package v1alpha1

//
//// Kluster represents a custom resource definition
//import (
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//// +genclient
//// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//type Kluster struct {
//	metav1.TypeMeta   `json:",inline"`
//	metav1.ObjectMeta `json:"metadata,omitempty"`
//
//	Spec KlusterSpec `json:"spec,omitempty"`
//}
//
//// KlusterList is a list of Kluster resources
//// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//type KlusterList struct {
//	metav1.TypeMeta `json:",inline"`
//	metav1.ListMeta `json:"metadata,omitempty"`
//	Items           []Kluster `json:"items"`
//}
//
//// KlusterSpec defines the desired state of Kluster
//type KlusterSpec struct {
//	// Define your specification fields here
//	Field1 string `json:"field1,omitempty"`
//	Field2 int    `json:"field2,omitempty"` // Example of an additional field
//}
