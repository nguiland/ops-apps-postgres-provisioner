package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Provisioner struct {
	db *sql.DB
}

func Connect(ctx context.Context, config ConnectionConfig) (*Provisioner, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.Username, config.Password),
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port))),
		Path:   "postgres",
	}

	q := u.Query()
	q.Set("sslmode", config.SSLMode)
	u.RawQuery = q.Encode()

	db, err := sql.Open("pgx", u.String())
	if err != nil {
		return nil, fmt.Errorf("Could not create PostgreSQL connection: %w", err)
	}

	db.SetMaxOpenConns(2)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("Could not connect to PostgreSQL: %w", err)
	}

	return &Provisioner{db: db}, nil
}

func (p *Provisioner) Close() error {
	return p.db.Close()
}

// Provision creates or updates the login role, creates the database if absent,
// and ensures the role owns it.
func (p *Provisioner) Provision(
	ctx context.Context,
	config DatabaseConfig,
	updatePassword bool,
) error {
	if err := p.ensureRole(ctx, config, updatePassword); err != nil {
		return err
	}

	if err := p.ensureDatabase(ctx, config); err != nil {
		return err
	}

	return p.ensureDatabaseOwner(ctx, config)
}

func (p *Provisioner) ensureRole(
	ctx context.Context,
	config DatabaseConfig,
	updatePassword bool,
) error {
	var exists bool

	if err := p.db.QueryRowContext(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)",
		config.Username,
	).Scan(&exists); err != nil {
		return fmt.Errorf("Could not check PostgreSQL role %q: %w", config.Username, err)
	}

	if exists && !updatePassword {
		return nil
	}

	var operation string

	if exists {
		operation = "ALTER ROLE %I LOGIN PASSWORD %L"
	} else {
		operation = "CREATE ROLE %I LOGIN PASSWORD %L"
	}

	var statement string

	if err := p.db.QueryRowContext(
		ctx,
		"SELECT format('"+operation+"', $1::text, $2::text)",
		config.Username,
		config.Password,
	).Scan(&statement); err != nil {
		return fmt.Errorf("Could not prepare role creation or update for %q: %w", config.Username, err)
	}

	if _, err := p.db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("Could not create or update PostgreSQL role %q: %w", config.Username, err)
	}

	return nil
}

func (p *Provisioner) ensureDatabase(
	ctx context.Context,
	config DatabaseConfig,
) error {
	var exists bool

	if err := p.db.QueryRowContext(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1::text)",
		config.Name,
	).Scan(&exists); err != nil {
		return fmt.Errorf("Could not check whether database %q exists: %w", config.Name, err)
	}

	if exists {
		return nil
	}

	var statement string

	if err := p.db.QueryRowContext(
		ctx,
		"SELECT format('CREATE DATABASE %I OWNER %I', $1::text, $2::text)",
		config.Name,
		config.Username,
	).Scan(&statement); err != nil {
		return fmt.Errorf("Could not prepare database creation for %q: %w", config.Name, err)
	}

	if _, err := p.db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("Could not create PostgreSQL database %q: %w", config.Name, err)
	}

	return nil
}

func (p *Provisioner) ensureDatabaseOwner(
	ctx context.Context,
	config DatabaseConfig,
) error {
	var statement string

	if err := p.db.QueryRowContext(
		ctx,
		"SELECT format('ALTER DATABASE %I OWNER TO %I', $1::text, $2::text)",
		config.Name,
		config.Username,
	).Scan(&statement); err != nil {
		return fmt.Errorf("Could not prepare ownership change for database %q and role %q: %w", config.Name, config.Username, err)
	}

	if _, err := p.db.ExecContext(ctx, statement); err != nil {
		return fmt.Errorf("Could not set role %q as owner of database %q: %w", config.Username, config.Name, err)
	}

	return nil
}
