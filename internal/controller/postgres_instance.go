package controller

import (
	"context"
	"fmt"
	"strings"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	"github.com/nguiland/ops-apps-postgres-provisioner/internal/postgres"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *PostgresReconciler) getInstance(
	ctx context.Context,
	req reconcile.Request,
) (*pgprovisioner.PostgresInstance, error) {
	instance := &pgprovisioner.PostgresInstance{}

	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	if instance.Namespace != r.Namespace {
		return nil, nil
	}

	return instance, nil
}

func (r *PostgresReconciler) getAdminCredentials(
	ctx context.Context,
	instance *pgprovisioner.PostgresInstance,
) (username, password string, err error) {
	adminNamespace := strings.TrimSpace(instance.Spec.AdminSecretRef.Namespace)
	if adminNamespace == "" {
		adminNamespace = instance.Namespace
	}

	adminSecret := &corev1.Secret{}
	secretKey := types.NamespacedName{
		Namespace: adminNamespace,
		Name:      instance.Spec.AdminSecretRef.Name,
	}

	if err := r.APIReader.Get(ctx, secretKey, adminSecret); err != nil {
		return "", "", fmt.Errorf(
			"read administrator Secret %s/%s: %w",
			adminNamespace,
			instance.Spec.AdminSecretRef.Name,
			err,
		)
	}

	_, username, password, err = requiredCredentials(adminSecret, false)
	if err != nil {
		return "", "", fmt.Errorf("administrator Secret: %w", err)
	}

	return username, password, nil
}

func (r *PostgresReconciler) connect(
	ctx context.Context,
	instance *pgprovisioner.PostgresInstance,
	adminUsername string,
	adminPassword string,
) (*postgres.Provisioner, error) {
	port := instance.Spec.Connection.Port

	if port == 0 {
		port = defaultPostgresPort
	}

	if port < 1 || port > 65535 {
		return nil, fmt.Errorf(
			"invalid PostgreSQL port %d: must be between 1 and 65535",
			port,
		)
	}

	sslMode := strings.TrimSpace(instance.Spec.Connection.SSLMode)
	if sslMode == "" {
		sslMode = defaultPostgresSSLMode
	}

	provisioner, err := postgres.Connect(ctx, postgres.ConnectionConfig{
		Host:          instance.Spec.Connection.Host,
		Port:          port,
		SSLMode:       sslMode,
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	return provisioner, nil
}
