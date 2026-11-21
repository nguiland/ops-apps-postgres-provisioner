package controller

import (
	"context"
	"fmt"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	"github.com/nguiland/ops-apps-postgres-provisioner/internal/postgres"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func databaseStatusesByKey(
	statuses []pgprovisioner.DatabaseStatus,
) map[string]pgprovisioner.DatabaseStatus {
	statusesByKey := make(map[string]pgprovisioner.DatabaseStatus, len(statuses))

	for _, status := range statuses {
		key := status.Namespace + "/" + status.SecretName
		statusesByKey[key] = status
	}

	return statusesByKey
}

func reconcileDatabaseSecrets(
	ctx context.Context,
	provisioner databaseProvisioner,
	secrets []corev1.Secret,
	previousStatuses map[string]pgprovisioner.DatabaseStatus,
) ([]pgprovisioner.DatabaseStatus, int) {
	logger := log.FromContext(ctx)

	nextStatuses := make([]pgprovisioner.DatabaseStatus, 0, len(secrets))
	failures := 0

	for i := range secrets {
		secret := &secrets[i]
		key := secret.Namespace + "/" + secret.Name

		database, username, password, err := requiredCredentials(secret, true)
		if err != nil {
			err = fmt.Errorf("Database credentials in Secret %s are invalid: %w", key, err)
		}

		previousStatus := previousStatuses[key]

		if err == nil {
			updatePassword := previousStatus.ObservedResourceVersion != secret.ResourceVersion

			err = provisioner.Provision(
				ctx,
				postgres.DatabaseConfig{
					Name:     database,
					Username: username,
					Password: password,
				},
				updatePassword,
			)
			if err != nil {
				err = fmt.Errorf(
					"Could not provision database %q for Secret %s: %w",
					database,
					key,
					err,
				)
			}
		}

		status := pgprovisioner.DatabaseStatus{
			Namespace:  secret.Namespace,
			SecretName: secret.Name,
			Database:   database,
			Username:   username,
		}

		if err != nil {
			status.ObservedResourceVersion = previousStatus.ObservedResourceVersion
			status.Ready = false
			status.Message = err.Error()
			failures++

			logger.Error(err, "Failed to reconcile managed database Secret", "secret", key)
		} else {
			status.ObservedResourceVersion = secret.ResourceVersion
			status.Ready = true
			status.Message = "Database and role are reconciled"
		}

		nextStatuses = append(nextStatuses, status)
	}

	return nextStatuses, failures
}
