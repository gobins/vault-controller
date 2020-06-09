package controllers

import (
	"context"

	apiv1 "github.com/gobins/vault-controller/api/v1"
)

func (r *PolicyReconciler) addFinalizer(instance *apiv1.Policy) error {
	instance.AddFinalizer(apiv1.PolicyFinalizer)
	return r.Update(context.Background(), instance)
}

func (r *PolicyReconciler) handleFinalizer(s *apiv1.Policy) error {
	if !s.HasFinalizer(apiv1.PolicyFinalizer) {
		return nil
	}

	if err := r.delete(s); err != nil {
		return err
	}
	s.RemoveFinalizer(apiv1.PolicyFinalizer)
	return r.Update(context.Background(), s)
}
