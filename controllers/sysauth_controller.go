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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	return ctrl.Result{}, nil
}

func (r *SysAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.SysAuth{}).
		Complete(r)
}
