# Overpacked App Helm Chart

Helm chart for deploying Overpacked App with either:
- CloudNativePG as the default in-cluster database option
- An external PostgreSQL database when desired

## Layout

- `Chart.yaml` - chart metadata and dependencies
- `values.yaml` - default production-safe values
- `values.schema.json` - JSON Schema validation for values (types, enums, JWT strength)
- `values-minimal.yaml` - minimal configuration for local/small self-hosted use with CloudNativePG cluster
- `values-minimal-external-db.yaml` - minimal configuration for local/small self-hosted use with external database mode
- `templates/` - Kubernetes manifests for app components and CloudNativePG resources

## Prerequisites

### Default mode: CloudNativePG

The chart defaults to `database.mode=cloudnativepg` and can install the CloudNativePG operator as an optional Helm dependency.

If you keep the operator enabled, Helm will install the operator dependency from the CloudNativePG chart repository:

- Repository: `https://cloudnative-pg.github.io/charts`
- Dependency: `cloudnative-pg`
- Alias: `cloudnativePgOperator`

The chart also creates a `Cluster` custom resource for the database.

### External database mode

If you want to use a managed PostgreSQL service or an existing database cluster, set:

- `database.mode=external`
- `cloudnativePgOperator.enabled=false`
- `database.external.host=<database host>`
- `database.external.port=<database port>`
- `database.external.sslmode=<ssl mode>`
- `database.auth.username=<database user>`
- `database.auth.password=<database password>`
- `database.auth.database=<database name>`

In external mode, the backend still receives its `DATABASE_URL` from the chart-managed secret, but the chart builds the connection string from the building blocks above.

## Required secrets

- `backend.auth.jwtSecret`

and either

- `backend.existingSecret`

or

- `backend.auth.password`, 
- `database.auth.password`
- `database.superuser.password` (CloudNativePG mode only)

## Install

### Minimal install

Update the secrets in values-minimal.yaml

```bash
helm dependency update deployment/helm/overpacked-app
helm upgrade --install overpacked-app deployment/helm/overpacked-app \
  -f deployment/helm/overpacked-app/values-minimal.yaml \
  --namespace overpacked \
  --create-namespace
```

### Install with an external PostgreSQL database

Update the secrets in values-minimal-external-db.yaml

```bash
helm dependency update deployment/helm/overpacked-app
helm upgrade --install overpacked-app deployment/helm/overpacked-app \
  -f deployment/helm/overpacked-app/values-minimal-external-db.yaml \
  --namespace overpacked \
  --create-namespace
```

## Notes

- The frontend is served through Nginx and proxies `/api` requests to the backend service.
- The default frontend host in the sample values is `overpacked.example.com`; change it for your environment.
- Credentials are required (see [Required secrets](#required-secrets)); the chart has no built-in defaults.
- Server-side backups are written to a PersistentVolumeClaim (`backend.backup.*`); set `backend.backup.enabled=false` to disable.
- If you disable the operator dependency in CloudNativePG mode, make sure the CloudNativePG operator and CRDs already exist in the target cluster.
