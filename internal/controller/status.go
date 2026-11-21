package controller

import (
	"context"
	"fmt"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *PostgresReconciler) updateStatus(
	ctx context.Context,
	instance *pgprovisioner.PostgresInstance,
	databases []pgprovisioner.DatabaseStatus,
	failures int,
) error {
	instance.Status.Databases = databases
	instance.Status.ObservedGeneration = instance.Generation

	condition := metav1.Condition{
		Type:               readyCondition,
		Status:             metav1.ConditionTrue,
		Reason:             reconciliationSucceeded,
		Message:            fmt.Sprintf("All %d managed database Secrets were reconciled successfully", len(databases)),
		ObservedGeneration: instance.Generation,
	}

	if failures > 0 {
		condition.Status = metav1.ConditionFalse
		condition.Reason = databaseProvisioningFailed
		condition.Message = fmt.Sprintf(
			"Could not reconcile %d of %d managed database Secrets; the remaining databases were reconciled successfully",
			failures,
			len(databases),
		)
	}

	meta.SetStatusCondition(&instance.Status.Conditions, condition)

	if err := r.Status().Update(ctx, instance); err != nil {
		return fmt.Errorf("Could not update status for PostgresInstance %s/%s: %w", instance.Namespace, instance.Name, err)
	}

	return nil
}

func (r *PostgresReconciler) setFailure(
	ctx context.Context,
	instance *pgprovisioner.PostgresInstance,
	cause error,
) (reconcile.Result, error) {
	meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
		Type:               readyCondition,
		Status:             metav1.ConditionFalse,
		Reason:             reconciliationFailed,
		Message:            cause.Error(),
		ObservedGeneration: instance.Generation,
	})

	instance.Status.ObservedGeneration = instance.Generation

	if err := r.Status().Update(ctx, instance); err != nil {
		return reconcile.Result{}, fmt.Errorf(
			"Could not update status for PostgresInstance %s/%s after reconciliation failed: %w",
			instance.Namespace,
			instance.Name,
			err,
		)
	}

	return reconcile.Result{RequeueAfter: retryInterval}, nil
}
