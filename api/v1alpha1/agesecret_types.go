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

// api/v1alpha1/agesecret_types.go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgeSecretTemplate defines Secret template settings (you currently use only .type).
type AgeSecretTemplate struct {
	// Default to Opaque if not specified.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=Opaque
	Type string `json:"type,omitempty"`
}

// AgeSecretSpec defines the desired state of the AgeSecret resource.
type AgeSecretSpec struct {
	// REQUIRED: Encrypted data (AGE armored or binary); key = Secret field name.
	// +kubebuilder:validation:Required
	EncryptedData map[string]string `json:"encryptedData"`

	// Secret template (e.g., Type: Opaque).
	// +kubebuilder:validation:Optional
	Template AgeSecretTemplate `json:"template,omitempty"`

	// Optional: list of recipients.
	// +kubebuilder:validation:Optional
	Recipients []string `json:"recipients,omitempty"`
}

// AgeSecretStatus defines observed state and metadata for the AgeSecret resource.
type AgeSecretStatus struct {
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Use ShortName = sea (as requested).
// +kubebuilder:resource:path=agesecrets,scope=Namespaced,shortName=sea
// Printer columns (shown in `kubectl get agesecrets`).
// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=`.status.secretName`
// +kubebuilder:printcolumn:name="Age",type=integer,JSONPath=`.metadata.generation`
type AgeSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgeSecretSpec   `json:"spec,omitempty"`
	Status AgeSecretStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type AgeSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgeSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgeSecret{}, &AgeSecretList{})
}
