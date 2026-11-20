package config

import "testing"

func TestParse_ArgsIsNil_ReturnDefaults(t *testing.T) {
	t.Setenv(podNamespaceEnv, "postgres-system")

	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}

	if want, got := ":8080", cfg.MetricsBindAddress; got != want {
		t.Errorf("MetricsBindAddress = %q, want %q", got, want)
	}

	if want, got := ":8081", cfg.HealthProbeBindAddress; got != want {
		t.Errorf("HealthProbeBindAddress = %q, want %q", got, want)
	}

	if want := true; cfg.LeaderElection != want {
		t.Errorf("LeaderElection = %t, want %t", cfg.LeaderElection, want)
	}

	if want, got := "postgres-system", cfg.Namespace; got != want {
		t.Errorf("Namespace = %q, want %q", got, want)
	}
}

func TestParse_NamespaceFlagIsProvided_OverrideEnvironmentNamespace(t *testing.T) {
	t.Setenv(podNamespaceEnv, "default")

	cfg, err := Parse([]string{"--namespace", "postgres-system"})
	if err != nil {
		t.Fatalf("Parse() returned an error: %v", err)
	}

	if want, got := "postgres-system", cfg.Namespace; got != want {
		t.Errorf("Namespace = %q, want %q", got, want)
	}
}

func TestParse_NamespaceIsBlank_ReturnError(t *testing.T) {
	t.Setenv(podNamespaceEnv, "   ")

	_, err := Parse(nil)
	if err == nil {
		t.Fatal("Parse() error = nil, want non-nil")
	}
}
