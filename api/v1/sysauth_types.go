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
	"fmt"

	"github.com/mitchellh/hashstructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//SysAuthFinalizer name of the sysauth finalizer
	SysAuthFinalizer = "sysauth.finalizers.vault.gobins.github.io"
	//WatchNamespace name of the namespace on which the controller is operating
	WatchNamespace = "vault-controller-system"
	//SysAuthFailedState state when failed
	SysAuthFailedState = "failed"
	//SysAuthCreatedState state when created
	SysAuthCreatedState = "created"
	//SysAuthUpdatedState state when updated
	SysAuthUpdatedState = "updated"
)

// SysAuthSpec defines the desired state of SysAuth
type SysAuthSpec struct {
	Path        string     `json:"path,omitempty"`
	Description string     `json:"description,omitempty"`
	Type        string     `json:"type,omitempty"`
	Local       bool       `json:"local,omitempty"`
	SealWrap    bool       `json:"seal_wrap,omitempty"`
	Config      AuthConfig `json:"config,omitempty"`
}

//AuthConfig define input config for SysAuth
type AuthConfig struct {
	DefaultLeaseTTL string `json:"default_lease_ttl,omitempty"`
	MaxLeaseTTL     string `json:"max_lease_ttl,omitempty"`
}

// SysAuthStatus defines the observed state of SysAuth
type SysAuthStatus struct {
	Hash  string `json:"hash,omitempty"`
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true

// SysAuth is the Schema for the sysauths API
type SysAuth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *SysAuthSpec   `json:"spec,omitempty"`
	Status *SysAuthStatus `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (s *SysAuth) IsBeingDeleted() bool {
	return !s.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsCreated returns true if a sysauth config has been created
func (s *SysAuth) IsCreated() bool {
	if s.Status == nil {
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

// GetHash returns a hash of the struct
func (s *SysAuth) GetHash() (string, error) {
	hash, err := hashstructure.Hash(s.Spec, nil)
	return fmt.Sprintf("%d", hash), err
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
