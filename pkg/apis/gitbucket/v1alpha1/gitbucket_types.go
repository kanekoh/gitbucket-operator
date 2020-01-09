package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitBucketSpec defines the desired state of GitBucket
type GitBucketSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:default := 1
	// +kubebuilder:validation:Minimum := 1
	Size int32 `json:"size"`

	// +kubebuilder:default := quay.io/hkaneko/gitbucket-docker:latest
	Image string `json:"image"`

	// +kubebuilder:default := false
	Enable_public bool `json:"enable_public"`

	// +kubebuilder:default := false
	Enable_database bool `json:"enable_database"`

	GitbucketHome *GitbucketHomeSpec `json:"gitbucketHome"`
}

type GitbucketHomeSpec struct {
	// +kubebuilder:default := true
	Ephemeral bool `json:"ephemeral"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum := 1
	Size int32 `json:"size"`

	// +kubebuilder:default := /opt/data/gitbucket
	MountPath string `json:"mount_path"`
}

// GitBucketStatus defines the observed state of GitBucket
type GitBucketStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitBucket is the Schema for the gitbuckets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=gitbuckets,scope=Namespaced
type GitBucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitBucketSpec   `json:"spec,omitempty"`
	Status GitBucketStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitBucketList contains a list of GitBucket
type GitBucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitBucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitBucket{}, &GitBucketList{})
}
