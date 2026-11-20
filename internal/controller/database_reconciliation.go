package controller

import (
	"context"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	"github.com/nguiland/ops-apps-postgres-provisioner/internal/postgres"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func databaseStatusesByKey(
	statuses []pgprovisioner.DatabaseStatus,
) map[string]pgprovisioner.DatabaseStatus {
	result := make(map[string]pgprovisioner.DatabaseStatus, len(statuses))

	for _, item := range statuses {
		key := item.Namespace + "/" + item.SecretName
		result[key] = item
	}

	return result
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

		item := pgprovisioner.DatabaseStatus{
			Namespace:  secret.Namespace,
			SecretName: secret.Name,
			Database:   database,
			Username:   username,
		}

		previous := previousStatuses[key]

		if err == nil {
			updatePassword := previous.ObservedResourceVersion != secret.ResourceVersion

			err = provisioner.Provision(
				ctx,
				postgres.DatabaseConfig{
					Name:     database,
					Username: username,
					Password: password,
				},
				updatePassword,
			)
		}

		if err != nil {
			item.ObservedResourceVersion = previous.ObservedResourceVersion
			item.Ready = false
			item.Message = err.Error()
			failures++

			logger.Error(err, "failed to provision database", "secret", key)
		} else {
			item.ObservedResourceVersion = secret.ResourceVersion
			item.Ready = true
			item.Message = "Database and role are reconciled"
		}

		nextStatuses = append(nextStatuses, item)
	}

	return nextStatuses, failures
}
