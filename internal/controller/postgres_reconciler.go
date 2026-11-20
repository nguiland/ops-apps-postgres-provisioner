package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type PostgresReconciler struct {
	client.Client
	APIReader client.Reader
	Namespace string
}

var _ reconcile.Reconciler = &PostgresReconciler{}

func (r *PostgresReconciler) Reconcile(
	ctx context.Context,
	req reconcile.Request,
) (reconcile.Result, error) {
	instance, err := r.getInstance(ctx, req)
	if err != nil {
		return reconcile.Result{}, err
	}
	if instance == nil {
		return reconcile.Result{}, nil
	}

	adminUser, adminPassword, err := r.getAdminCredentials(ctx, instance)
	if err != nil {
		return r.setFailure(ctx, instance, err)
	}

	provisioner, err := r.connect(ctx, instance, adminUser, adminPassword)
	if err != nil {
		return r.setFailure(ctx, instance, err)
	}
	defer provisioner.Close()

	databaseSecrets := newDatabaseSecrets(r.Client, r.Namespace)

	secrets, err := databaseSecrets.list(ctx, instance)
	if err != nil {
		return r.setFailure(ctx, instance, err)
	}

	previousStatuses := databaseStatusesByKey(instance.Status.Databases)

	nextStatuses, failures := reconcileDatabaseSecrets(
		ctx,
		provisioner,
		secrets,
		previousStatuses,
	)

	if err := r.updateStatus(ctx, instance, nextStatuses, failures); err != nil {
		return reconcile.Result{}, err
	}

	if failures > 0 {
		return reconcile.Result{RequeueAfter: retryInterval}, nil
	}

	return reconcile.Result{RequeueAfter: reconcileInterval}, nil
}
