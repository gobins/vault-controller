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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/gobins/vault-controller/api/v1"
	vaultapi "github.com/hashicorp/vault/api"
)

// SysAuthReconciler reconciles a SysAuth object
type SysAuthReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	APIClient *vaultapi.Client
	Recorder  record.EventRecorder
}

// +kubebuilder:rbac:groups=vault.gobins.github.io,resources=sysauths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.gobins.github.io,resources=sysauths/status,verbs=get;update;patch

func (r *SysAuthReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sysauth", req.NamespacedName)

	sysauth := &apiv1.SysAuth{}
	log.Info(fmt.Sprintf("starting reconcile loop for %v", req.NamespacedName))
	defer log.Info(fmt.Sprintf("completed reconcile loop for %v", req.NamespacedName))

	err := r.Get(ctx, req.NamespacedName, sysauth)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if sysauth.IsBeingDeleted() {
		log.Info("run finalizer")
		err := r.handleFinalizer(sysauth)
		if err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "deleting finalizer", fmt.Sprintf("failed to delete finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "Deleted", "Object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	if !sysauth.HasFinalizer(apiv1.SysAuthFinalizer) {
		r.Log.Info(fmt.Sprintf("add finalizer for %v", req.NamespacedName))
		if err := r.addFinalizer(sysauth); err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "adding finalizer", fmt.Sprintf("failed to add finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
		}
		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "added", "object finalizer is added")
		return ctrl.Result{}, nil
	}

	if !sysauth.IsCreated() {
		r.Log.Info(fmt.Sprintf("submit for %v", req.NamespacedName))
		if err := r.create(sysauth); err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "submitting object", fmt.Sprintf("failed to submit object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when creating sysauth: %v", err)
		}
		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "created", "sysauth is created")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SysAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.SysAuth{}).
		Complete(r)
}

func (r *SysAuthReconciler) delete(s *apiv1.SysAuth) error {
	r.Log.Info(fmt.Sprintf("deleting sysauth %s", s.GetName()))

	if s.Status.CreatedTimestamp == nil {
		return nil
	}
	return r.APIClient.Sys().DisableAuth(s.Spec.Path)
}

func (r *SysAuthReconciler) create(s *apiv1.SysAuth) error {
	r.Log.Info(fmt.Sprintf("creating sysauth %s", s.GetName()))
	return r.APIClient.Sys().EnableAuthWithOptions(s.Spec.Path, s.Spec.Options)
}

// IsUptoDate returns true if a sysauth config is current
func (r *SysAuthReconciler) IsUptoDate(s *apiv1.SysAuth) (bool, error) {
	if s.Status.CreatedTimestamp == nil {
		return false, nil
	}
	_, err := r.APIClient.Sys().ListAuth()
	if err != nil {
		return false, fmt.Errorf("error when listing sysauth: %v", err)
	}
	// for _, auth := range authList{
	// 	auth.
	// }
	// s.Spec.Options.
	return true, nil
}
