# Postgres Provisioner

A Kubernetes controller that provisions PostgreSQL roles and databases from labeled Kubernetes Secrets.

The provisioner runs once per Kubernetes cluster and can manage multiple PostgreSQL instances and database Secrets across namespaces.

## Architecture

- Deploy the provisioner in a cluster management namespace, such as `local`, `int`, or `prod`.
- Create a `PostgresInstance` resource in the management namespace for each PostgreSQL server to be managed. For example, the resource named `main` represents the default PostgreSQL server.
- Each `PostgresInstance` references an administrator Secret containing the PostgreSQL administrator credentials. By default, this Secret is expected in the same namespace as the `PostgresInstance`.
- Tenant namespaces opt in to database provisioning by labeling their database Secrets. Capsule continues to govern which teams can create Secrets in their tenant namespaces; the provisioner does not install anything into tenant namespaces.
- Multiple database Secrets per tenant namespace are supported.
- The label `postgres-provisioner.io/instance` routes each database Secret to a named `PostgresInstance`.
- The administrator PostgreSQL account needs `CREATEROLE`, `CREATEDB`, and the ability to assign database ownership to the provisioned roles.
- Secret removal is non-destructive: the provisioner never drops PostgreSQL roles or databases.

### Secret conventions

The administrator Secret must contain:

| Key | Required |
|---|---|
| `username` | Yes |
| `password` | Yes |

Each managed database Secret must contain:

| Key | Required |
|---|---|
| `username` | Yes |
| `password` | Yes |
| `database` | No |

When `database` is omitted, its value defaults to `username`.

A database Secret must also have these labels:

```yaml
postgres-provisioner.io/managed: "true"
postgres-provisioner.io/instance: main
```

## Repository structure

This repository contains the Go application, its Kubernetes API definitions, and the generated Kubernetes manifests.

The Helm chart is maintained separately in its own Git repository.

The CRD definition is maintained here as the authoritative source and generated from the Go API types. The generated CRD is also distributed through the Helm chart repository for installation.

## Development

### Prerequisites

- Go
- Docker with Buildx
- Access to a Kubernetes cluster for deployment and integration testing

### Generate manifests

Generate the DeepCopy methods:

```sh
make generate
```

Generate the CRD and RBAC manifests:

```sh
make manifests
```

The generated files are written to:

- `config/crd/bases/`: CustomResourceDefinition manifests.
- `config/rbac/`: Kubernetes RBAC manifests.

RBAC requirements are declared using Kubebuilder markers in `internal/controller/doc.go`.

### Test and build

```sh
make test
make build
```

### Build and publish the container image

```sh
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/nguiland/postgres-provisioner:0.1.0 --push .
```

### Go module

The Go module is:

```text
github.com/nguiland/ops-apps-postgres-provisioner
```

## Deployment

The provisioner is deployed using the Helm chart maintained in a separate repository.

The chart packages the application image, Kubernetes resources, RBAC configuration, and CRD for installation.

The CRD is maintained in this repository and must be synchronized with the chart repository when a new application release introduces API changes.

**CRD lifecycle:** Helm installs CRDs from its `crds/` directory on initial installation but does not upgrade or delete them. Apply CRD schema changes separately when upgrading the provisioner.

See the Helm chart repository for installation instructions and configuration options.

## Important security considerations

The provisioner must watch Secrets across namespaces to discover tenant databases.

Kubernetes RBAC does not support label-filtered `list` or `watch` permissions. Although the provisioner's informer cache filters Secrets to those opted in for management, its ServiceAccount still requires cluster-wide `get`, `list`, and `watch` permissions on Secrets.

Consequently, restrict who can:

- Create or update `PostgresInstance` resources.
- Create or update Secrets with the provisioner's opt-in labels.
- Modify the labels that route database Secrets to PostgreSQL instances.

These permissions are important because anyone able to opt a Secret into management can request database and role provisioning using the credentials supplied in that Secret.