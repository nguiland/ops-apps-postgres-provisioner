package controller

import (
	"context"

	"github.com/nguiland/ops-apps-postgres-provisioner/internal/postgres"
)

// databaseProvisioner describes the PostgreSQL operation required by
// the controller. The concrete implementation is provided by the
// postgres package.
type databaseProvisioner interface {
	Provision(ctx context.Context, config postgres.DatabaseConfig, updatePassword bool) error
}
