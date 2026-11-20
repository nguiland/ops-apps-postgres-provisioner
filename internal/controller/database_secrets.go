package controller

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pgprovisioner "github.com/nguiland/ops-apps-postgres-provisioner/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type databaseSecrets struct {
	reader    client.Reader
	namespace string
}

func newDatabaseSecrets(reader client.Reader, namespace string) databaseSecrets {
	return databaseSecrets{
		reader:    reader,
		namespace: namespace,
	}
}

func (s databaseSecrets) list(
	ctx context.Context,
	instance *pgprovisioner.PostgresInstance,
) ([]corev1.Secret, error) {
	secrets := &corev1.SecretList{}

	selector := client.MatchingLabels{
		ManagedLabel:  "true",
		instanceLabel: instance.Name,
	}

	if err := s.reader.List(ctx, secrets, selector); err != nil {
		return nil, fmt.Errorf("list managed database Secrets: %w", err)
	}

	sort.Slice(secrets.Items, func(i, j int) bool {
		a, b := secrets.Items[i], secrets.Items[j]

		if a.Namespace == b.Namespace {
			return a.Name < b.Name
		}

		return a.Namespace < b.Namespace
	})

	return secrets.Items, nil
}

func (s databaseSecrets) toInstance(
	_ context.Context,
	obj client.Object,
) []reconcile.Request {
	secret, ok := obj.(*corev1.Secret)
	if !ok || secret.Labels[ManagedLabel] != "true" {
		return nil
	}

	instance := strings.TrimSpace(secret.Labels[instanceLabel])
	if instance == "" {
		return nil
	}

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: s.namespace,
				Name:      instance,
			},
		},
	}
}
