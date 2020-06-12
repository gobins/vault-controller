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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/gobins/vault-controller/api/v1"
	vaultv1 "github.com/gobins/vault-controller/api/v1"
)

// PolicyReconciler reconciles a Policy object
type PolicyReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	APIClient *vaultapi.Client
	Recorder  record.EventRecorder
}

// +kubebuilder:rbac:groups=vault.gobins.github.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.gobins.github.io,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create

func (r *PolicyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("policy", req.NamespacedName)

	policy := &apiv1.Policy{}
	log.Info(fmt.Sprintf("starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("completed reconcile loop for %v", req.NamespacedName))
	err := r.Get(ctx, req.NamespacedName, policy)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	// Initializing vault config
	config, err := r.getConfig()
	if err != nil {
		r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to get vault config: %s", err))
		return ctrl.Result{}, nil
	}
	if config != nil {
		address := config.Data["address"]
		token := config.Data["token"]
		r.APIClient, err = GetClient(address, token)
	}
	if err != nil {
		r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to init vault client: %s", err))
		return ctrl.Result{}, nil
	}

	if policy.IsBeingDeleted() {
		log.Info("run finalizer")
		err := r.handleFinalizer(policy)
		if err != nil {
			r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to delete finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(policy, corev1.EventTypeNormal, "deleted", "object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	isUptoDate, err := r.IsUptoDate(policy)
	if err != nil {
		r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to check object upto date: %s", err))
		return ctrl.Result{}, fmt.Errorf("error when checking policy IsUptoDate: %v", err)
	}

	if !policy.IsCreated() || !isUptoDate {
		r.Log.Info(fmt.Sprintf("creating/updating policy %v", policy.Spec.Name))
		if err := r.put(policy); err != nil {
			if !policy.IsCreated() {
				r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to create object: %s", err))
			}
			r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to update object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when creating policy: %v", err)
		}

		if !policy.HasFinalizer(apiv1.PolicyFinalizer) {
			r.Log.Info(fmt.Sprintf("add finalizer for %v", req.NamespacedName))
			if err := r.addFinalizer(policy); err != nil {
				r.Recorder.Event(policy, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to add finalizer: %s", err))
				return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
			}
			r.Recorder.Event(policy, corev1.EventTypeNormal, "added", "object finalizer is added")
		}
		if !policy.IsCreated() {
			r.Recorder.Event(policy, corev1.EventTypeNormal, "created", "policy is created")
		}
		r.Recorder.Event(policy, corev1.EventTypeNormal, "updated", "policy is updated")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *PolicyReconciler) getConfig() (*corev1.ConfigMap, error) {
	config := &corev1.ConfigMap{}
	err := r.Client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      "config",
			Namespace: apiv1.WatchNamespace,
		},
		config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (r *PolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1.Policy{}).
		Complete(r)
}

func (r *PolicyReconciler) delete(p *apiv1.Policy) error {
	r.Log.Info(fmt.Sprintf("deleting policy %s", p.GetName()))
	if p.Status == nil {
		return nil
	}
	return r.APIClient.Sys().DeletePolicy(p.Spec.Name)
}

func (r *PolicyReconciler) put(p *apiv1.Policy) error {
	err := r.APIClient.Sys().PutPolicy(p.Spec.Name, p.Spec.Rules)
	if err != nil {
		return err
	}
	hash, err := p.GetHash()
	if err != nil {
		return err
	}
	p.Status = &apiv1.PolicyStatus{
		Hash:  hash,
		State: apiv1.PolicyCreatedState,
	}
	err = r.Update(context.Background(), p)
	if err != nil {
		return err
	}
	return nil
}

// IsUptoDate returns true if a sysauth config is current
func (p *PolicyReconciler) IsUptoDate(s *apiv1.Policy) (bool, error) {
	hash, err := s.GetHash()
	if err != nil {
		return false, fmt.Errorf("error when calculating policy hash: %v", err)
	}
	if s.Status == nil {
		return false, nil
	}
	if s.Status.Hash != hash {
		return false, nil
	}
	return true, nil
}
