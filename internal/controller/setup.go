package controller

import (
	"strings"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *PostgresReconciler) SetupWithManager(mgr ctrl.Manager) error {
	databaseSecrets := newDatabaseSecrets(r.Client, r.Namespace)

	return ctrl.NewControllerManagedBy(mgr).
		For(&pgprovisioner.PostgresInstance{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(databaseSecrets.toInstance),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
				labels := obj.GetLabels()
				return labels[ManagedLabel] == "true" && strings.TrimSpace(labels[instanceLabel]) != ""
			})),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		Complete(r)
}
