package main

import (
	"fmt"
	"os"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	"github.com/nguiland/ops-apps-postgres-provisioner/internal/config"
	"github.com/nguiland/ops-apps-postgres-provisioner/internal/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&cfg.Zap)))
	klog.SetLogger(log.Log.WithName("klog"))

	scheme, err := registerTypes()
	if err != nil {
		return fmt.Errorf("register Kubernetes API types: %w", err)
	}

	cacheOptions := configureCache(cfg.Namespace)

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("get Kubernetes REST configuration: %w", err)
	}

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: cfg.MetricsBindAddress},
		HealthProbeBindAddress: cfg.HealthProbeBindAddress,
		LeaderElection:         cfg.LeaderElection,
		LeaderElectionID:       "leader.postgres-provisioner.io",
		Cache:                  cacheOptions,
	})
	if err != nil {
		return fmt.Errorf("create controller manager: %w", err)
	}

	reconciler := &controller.PostgresReconciler{
		Client:    mgr.GetClient(),
		APIReader: mgr.GetAPIReader(),
		Namespace: cfg.Namespace,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("set up Postgres reconciler: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("register health check: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("register readiness check: %w", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("start controller manager: %w", err)
	}

	return nil
}

func registerTypes() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("register core Kubernetes API types: %w", err)
	}

	if err := pgprovisioner.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("register Postgres API types: %w", err)
	}

	return scheme, nil
}

func configureCache(namespace string) cache.Options {
	return cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Secret{}: {
				Label: labels.Set{
					controller.ManagedLabel: "true",
				}.AsSelector(),
			},
			&pgprovisioner.PostgresInstance{}: {
				Namespaces: map[string]cache.Config{
					namespace: {},
				},
			},
		},
	}
}
