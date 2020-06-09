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
	"k8s.io/apimachinery/pkg/types"
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
	// Initializing vault config
	config, err := r.getConfig()
	if err != nil {
		r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to get vault config: %s", err))
		return ctrl.Result{}, nil
	}
	if config != nil {
		address := config.Data["address"]
		token := config.Data["token"]
		r.APIClient, err = GetClient(address, token)
	}
	if err != nil {
		r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to init vault client: %s", err))
		return ctrl.Result{}, nil
	}

	if sysauth.IsBeingDeleted() {
		log.Info("run finalizer")
		err := r.handleFinalizer(sysauth)
		if err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to delete finalizer: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when handling finalizer: %v", err)
		}
		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "deleted", "object finalizer is deleted")
		return ctrl.Result{}, nil
	}

	isUptoDate, err := r.IsUptoDate(sysauth)
	if err != nil {
		r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to check object upto date: %s", err))
		return ctrl.Result{}, fmt.Errorf("error when checking sysauth IsUptoDate: %v", err)
	}

	if !sysauth.IsCreated() {
		r.Log.Info(fmt.Sprintf("creating sysauth %v", sysauth.Spec.Path))
		if err := r.create(sysauth); err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to creating object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when creating sysauth: %v", err)
		}

		if !sysauth.HasFinalizer(apiv1.SysAuthFinalizer) {
			r.Log.Info(fmt.Sprintf("add finalizer for %v", req.NamespacedName))
			if err := r.addFinalizer(sysauth); err != nil {
				r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to add finalizer: %s", err))
				return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
			}
			r.Recorder.Event(sysauth, corev1.EventTypeNormal, "added", "object finalizer is added")
			return ctrl.Result{}, nil
		}

		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "created", "sysauth is created")
		return ctrl.Result{}, nil
	}

	if !isUptoDate {
		r.Log.Info(fmt.Sprintf("updating sysauth %v", sysauth.Spec.Path))
		if err := r.update(sysauth); err != nil {
			r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to update object: %s", err))
			return ctrl.Result{}, fmt.Errorf("error when updating sysauth: %v", err)
		}

		if !sysauth.HasFinalizer(apiv1.SysAuthFinalizer) {
			r.Log.Info(fmt.Sprintf("add finalizer for %v", req.NamespacedName))
			if err := r.addFinalizer(sysauth); err != nil {
				r.Recorder.Event(sysauth, corev1.EventTypeWarning, "failed", fmt.Sprintf("failed to add finalizer: %s", err))
				return ctrl.Result{}, fmt.Errorf("error when adding finalizer: %v", err)
			}
			r.Recorder.Event(sysauth, corev1.EventTypeNormal, "added", "object finalizer is added")
			return ctrl.Result{}, nil
		}
		r.Recorder.Event(sysauth, corev1.EventTypeNormal, "updated", "sysauth is updated")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SysAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.SysAuth{}).
		Complete(r)
}

func (r *SysAuthReconciler) getConfig() (*corev1.ConfigMap, error) {
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

func (r *SysAuthReconciler) delete(s *apiv1.SysAuth) error {
	r.Log.Info(fmt.Sprintf("deleting sysauth %s", s.GetName()))

	if s.Status == nil {
		return nil
	}
	return r.APIClient.Sys().DisableAuth(s.Spec.Path)
}

func (r *SysAuthReconciler) create(s *apiv1.SysAuth) error {
	r.Log.Info(fmt.Sprintf("creating sysauth %s", s.GetName()))
	err := r.APIClient.Sys().EnableAuthWithOptions(s.Spec.Path,
		&vaultapi.MountInput{
			Description: s.Spec.Description,
			Type:        s.Spec.Type,
			Local:       s.Spec.Local,
			SealWrap:    s.Spec.SealWrap,
			Config: vaultapi.MountConfigInput{
				DefaultLeaseTTL: s.Spec.Config.DefaultLeaseTTL,
				MaxLeaseTTL:     s.Spec.Config.MaxLeaseTTL,
			},
		})
	if err != nil {
		return err
	}
	hash, err := s.GetHash()
	if err != nil {
		return err
	}
	s.Status = &apiv1.SysAuthStatus{
		Hash:   hash,
		Status: apiv1.SysAuthCreatedState,
	}
	err = r.Update(context.Background(), s)
	if err != nil {
		return err
	}
	return nil
}

func (r *SysAuthReconciler) update(s *apiv1.SysAuth) error {
	r.Log.Info(fmt.Sprintf("creating sysauth %s", s.GetName()))
	err := r.APIClient.Sys().TuneMount("/auth/"+s.Spec.Path,
		vaultapi.MountConfigInput{
			Description:     &s.Spec.Description,
			DefaultLeaseTTL: s.Spec.Config.DefaultLeaseTTL,
			MaxLeaseTTL:     s.Spec.Config.MaxLeaseTTL,
		})
	if err != nil {
		return err
	}
	hash, err := s.GetHash()
	if err != nil {
		return err
	}
	s.Status = &apiv1.SysAuthStatus{
		Hash:   hash,
		Status: apiv1.SysAuthUpdatedState,
	}
	err = r.Update(context.Background(), s)
	if err != nil {
		return err
	}
	return nil
}

// IsUptoDate returns true if a sysauth config is current
func (r *SysAuthReconciler) IsUptoDate(s *apiv1.SysAuth) (bool, error) {
	hash, err := s.GetHash()
	if err != nil {
		return false, fmt.Errorf("error when calculating sysauth hash: %v", err)
	}
	if s.Status == nil {
		return false, nil
	}
	if s.Status.Hash != hash {
		return false, nil
	}
	return true, nil
}
