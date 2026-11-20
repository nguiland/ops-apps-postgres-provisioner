package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const podNamespaceEnv = "POD_NAMESPACE"

type Config struct {
	MetricsBindAddress     string
	HealthProbeBindAddress string
	LeaderElection         bool
	Namespace              string
	Zap                    zap.Options
}

func Parse(args []string) (Config, error) {
	var cfg Config

	flags := flag.NewFlagSet("postgres-provisioner", flag.ContinueOnError)

	flags.StringVar(
		&cfg.MetricsBindAddress,
		"metrics-bind-address",
		":8080",
		"Metrics endpoint address.",
	)

	flags.StringVar(
		&cfg.HealthProbeBindAddress,
		"health-probe-bind-address",
		":8081",
		"Health probe endpoint address.",
	)

	flags.BoolVar(
		&cfg.LeaderElection,
		"leader-elect",
		true,
		"Enable leader election.",
	)

	flags.StringVar(
		&cfg.Namespace,
		"namespace",
		os.Getenv(podNamespaceEnv),
		"Namespace containing PostgresInstance resources.",
	)

	cfg.Zap = zap.Options{Development: false}
	cfg.Zap.BindFlags(flags)

	if err := flags.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	cfg.Namespace = strings.TrimSpace(cfg.Namespace)

	if cfg.Namespace == "" {
		return Config{}, errors.New("namespace must be provided")
	}

	return cfg, nil
}
