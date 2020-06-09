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
	//PolicyFinalizer name of the sysauth finalizer
	PolicyFinalizer = "sysauth.finalizers.vault.gobins.github.io"
	//PolicyWatchNamespace name of the namespace on which the controller is operating
	PolicyWatchNamespace = "vault-controller-system"
	//PolicyFailedState state when failed
	PolicyFailedState = "failed"
	//PolicyCreatedState state when created
	PolicyCreatedState = "created"
	//PolicyUpdatedState state when updated
	PolicyUpdatedState = "updated"
)

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//Name is the policy name
	Name string `json:"name,omitempty"`
	//Rules defines the vault policy rules
	Rules string `json:"rules,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State string `json:"state,omitempty"`
	Hash  string `json:"hash,omitempty"`
}

// +kubebuilder:object:root=true

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *PolicySpec   `json:"spec,omitempty"`
	Status *PolicyStatus `json:"status,omitempty"`
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (p *Policy) IsBeingDeleted() bool {
	return !p.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsCreated returns true if a sysauth config has been created
func (p *Policy) IsCreated() bool {
	if p.Status == nil {
		return false
	}
	return true
}

// HasFinalizer returns true if item has a finalizer with input name
func (p *Policy) HasFinalizer(name string) bool {
	return containsString(p.ObjectMeta.Finalizers, name)
}

// AddFinalizer adds the input finalizer
func (p *Policy) AddFinalizer(name string) {
	p.ObjectMeta.Finalizers = append(p.ObjectMeta.Finalizers, name)
}

// RemoveFinalizer removes the input finalizer
func (p *Policy) RemoveFinalizer(name string) {
	p.ObjectMeta.Finalizers = removeString(p.ObjectMeta.Finalizers, name)
}

// GetHash returns a hash of the struct
func (p *Policy) GetHash() (string, error) {
	hash, err := hashstructure.Hash(p.Spec.Rules, nil)
	return fmt.Sprintf("%d", hash), err
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}
