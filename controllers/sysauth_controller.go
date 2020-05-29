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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	vaultv1 "github.com/gobins/vault-controller/api/v1"
	"github.com/gobins/vault-controller/controllers/vault"
)

// SysAuthReconciler reconciles a SysAuth object
type SysAuthReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vault.vault.gobins.io,resources=sysauths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vault.vault.gobins.io,resources=sysauths/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vault.vault.gobins.io,resources=configs,verbs=get;list;watch

func (r *SysAuthReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("sysauth", req.NamespacedName)

	// your logic here
	var auth vaultv1.SysAuth
	var config vaultv1.Config
	err := r.Get(ctx, req.NamespacedName, &auth)
	if err != nil {
		log.Error(err, "enable to fetch sysauth")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.Get(ctx, types.NamespacedName{Name: "config-sample", Namespace: "vault-controller-system"}, &config)
	if err != nil {
		log.Error(err, "unable to fetch config")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if auth.Status.Updated != "True" {
		c := vault.GetClient(config.Spec.Url, config.Spec.Token)
		log.Info("setting vault client with url %s and token %s", config.Spec.Url, config.Spec.Token)
		err := c.Sys().EnableAuth(auth.Spec.Path, auth.Spec.Type, auth.Spec.Description)
		if err != nil {
			log.Error(err, "error authenticating to vault")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		auth.Status.Updated = "True"
		err = r.Update(ctx, &auth)
		if err != nil {
			log.Error(err, "unable to update auth")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		log.Info("status updated to True")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *SysAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vaultv1.SysAuth{}).
		Complete(r)
}
