/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SysAuthSpec defines the desired state of SysAuth
type SysAuthSpec struct {
	Path        string `json:"path,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// SysAuthStatus defines the observed state of SysAuth
type SysAuthStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CreatedTimestamp *metav1.Timestamp `json:"createdTimestamp,omitempty"`
	UpdatedTimestamp *metav1.Timestamp `json:"updatedTimestamp,omitempty"`
}

// +kubebuilder:object:root=true

// SysAuth is the Schema for the sysauths API
type SysAuth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SysAuthSpec   `json:"spec,omitempty"`
	Status SysAuthStatus `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (s *SysAuth) IsBeingDeleted() bool {
	return !s.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted returns true if a deletion timestamp is set
func (s *SysAuth) IsSubmitted() bool {
	if s.Status.CreatedTimestamp == nil {
		return false
	}
	return true
}

// +kubebuilder:object:root=true

// SysAuthList contains a list of SysAuth
type SysAuthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SysAuth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SysAuth{}, &SysAuthList{})
}
