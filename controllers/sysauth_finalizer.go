package controllers

import (
	"context"

	apiv1 "github.com/gobins/vault-controller/api/v1"
)

func (r *SysAuthReconciler) addFinalizer(instance *apiv1.SysAuth) error {
	instance.AddFinalizer(apiv1.SysAuthFinalizer)
	return r.Update(context.Background(), instance)
}

func (r *SysAuthReconciler) handleFinalizer(s *apiv1.SysAuth) error {
	if !s.HasFinalizer(apiv1.SysAuthFinalizer) {
		return nil
	}

	if err := r.delete(s); err != nil {
		return err
	}
	s.RemoveFinalizer(apiv1.SysAuthFinalizer)
	return r.Update(context.Background(), s)
}
