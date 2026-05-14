# Overpacked App Helm Chart

Production-grade Helm chart for deploying Overpacked App with either:
- CloudNativePG as the default in-cluster database option
- An external PostgreSQL database when desired

## Layout

- `Chart.yaml` - chart metadata and dependencies
- `values.yaml` - default production-safe values
- `values-production.yaml` - recommended production baseline
- `values-external-db.yaml` - example for external database mode
- `templates/` - Kubernetes manifests for app components and CNPG resources

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
- `database.external.username=<database user>`
- `database.external.password=<database password>`
- `database.external.database=<database name>`
- `database.external.sslmode=<ssl mode>`

In external mode, the backend still receives its `DATABASE_URL` from the chart-managed secret, but the chart builds the connection string from the building blocks above.

## Install

### Production install with CloudNativePG

```bash
helm dependency update deployment/helm/overpacked-app
helm upgrade --install overpacked-app deployment/helm/overpacked-app \
  -f deployment/helm/overpacked-app/values-production.yaml \
  --namespace overpacked \
  --create-namespace
```

### Install with an external PostgreSQL database

```bash
helm dependency update deployment/helm/overpacked-app
helm upgrade --install overpacked-app deployment/helm/overpacked-app \
  -f deployment/helm/overpacked-app/values-external-db.yaml \
  --namespace overpacked \
  --create-namespace
```

## Notes

- The frontend is served through Nginx and proxies `/api` requests to the backend service.
- The default frontend host in the sample values is `overpacked.example.com`; change it for your environment.
- Before production use, replace all example passwords and secrets with real values.
- If you disable the operator dependency in CloudNativePG mode, make sure the CloudNativePG operator and CRDs already exist in the target cluster.
