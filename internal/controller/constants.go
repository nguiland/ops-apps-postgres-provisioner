package controller

import "time"

const (
	// Kubernetes labels.
	ManagedLabel  = "postgres-provisioner.io/managed" // Exported because it is used in the main package
	instanceLabel = "postgres-provisioner.io/instance"

	// Database Secret data keys.
	databaseNameKey = "database"
	usernameKey     = "username"
	passwordKey     = "password"

	// Status conditions and reasons.
	readyCondition             = "Ready"
	reconciliationSucceeded    = "ReconciliationSucceeded"
	reconciliationFailed       = "ReconciliationFailed"
	databaseProvisioningFailed = "DatabaseProvisioningFailed"

	// PostgreSQL connection defaults.
	defaultPostgresPort    = 5432
	defaultPostgresSSLMode = "prefer"

	// Reconciliation intervals.
	retryInterval     = time.Minute
	reconcileInterval = 10 * time.Minute
)
