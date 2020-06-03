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
	vaultapi "github.com/hashicorp/vault/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//SysAuthFinalizer name of the sysauth finalizer
const SysAuthFinalizer = "sysauth.finalizers.vault.gobins.github.io"

// SysAuthSpec defines the desired state of SysAuth
type SysAuthSpec struct {
	Path        string                      `json:"path,omitempty"`
	Description string                      `json:"description,omitempty"`
	Type        string                      `json:"type,omitempty"`
	Options     *vaultapi.EnableAuthOptions `json:"options,omitempty"`
}

// SysAuthStatus defines the observed state of SysAuth
type SysAuthStatus struct {
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

// IsCreated returns true if a sysauth config has been created
func (s *SysAuth) IsCreated() bool {
	if s.Status.CreatedTimestamp == nil {
		return false
	}
	return true
}

// HasFinalizer returns true if item has a finalizer with input name
func (s *SysAuth) HasFinalizer(name string) bool {
	return containsString(s.ObjectMeta.Finalizers, name)
}

// AddFinalizer adds the input finalizer
func (s *SysAuth) AddFinalizer(name string) {
	s.ObjectMeta.Finalizers = append(s.ObjectMeta.Finalizers, name)
}

// RemoveFinalizer removes the input finalizer
func (s *SysAuth) RemoveFinalizer(name string) {
	s.ObjectMeta.Finalizers = removeString(s.ObjectMeta.Finalizers, name)
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
