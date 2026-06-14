# Dev Setup, Workflow & Git

## Prerequisites

- Go (version matching `backend/go.mod`)
- Node.js + npm (matching `frontend/package.json`)
- Docker (optional but recommended for the full stack)

## Setup & running

```bash
make install          # install backend + frontend deps
make up               # run the full dev stack (frontend, backend, db) via docker compose
make backend          # backend container only (with db)
make frontend         # frontend only
```

## Common commands

```bash
make test             # backend + frontend checks
make test-backend     # go test ./... (unit; integration tests self-skip without a DB)
make test-backend-container  # full Go suite incl. the full-stack E2E, against a containerized
                             # Postgres, with a coverage profile (backend/coverage.out)
make test-frontend    # vue-tsc + theme/icon lint
make build            # build backend + frontend
make gen-api-go       # regenerate Go API types from dev/openapi.yaml
make gen-api          # regenerate frontend OpenAPI types
make seed             # run database seeds (local)
```

## Fork-based git workflow (required)

Contributions go through forks, not direct branches on the main repo:

```bash
# Clone your fork
git clone git@github.com:<your-user>/overpacked-app.git
cd overpacked-app

# Add the main repo as upstream
git remote add upstream git@github.com:via-justa/overpacked-app.git

# Sync main
git fetch upstream
git checkout main
git merge --ff-only upstream/main

# Create a feature branch
git checkout -b feat/my-change
```

## Contribution workflow

1. Create an issue first for significant changes.
2. Fork, clone, add `upstream`.
3. Branch from an up-to-date `main`.
4. Keep PRs **focused and small**.
5. Run relevant tests before opening a PR.
6. Fill in the PR template completely.

## Commit & PR guidance

- Clear commit messages.
- PR titles should use release keywords where possible: `(feat)`, `(fix)`, `(chore)`. (These
  drive release-drafter labeling.)
- Branch names follow the type prefix, e.g. `feat/...`, `fix/...`, `chore/...`.
- Link the relevant issue in the PR description.
- Include screenshots for UI changes.
- Run theme lint for UI changes: `cd frontend && npm run lint:theme`.

## Security hygiene

- Never commit secrets, tokens, credentials, or private keys.
- Don't bake secrets into Docker images.
- Report security issues privately to maintainers rather than in public issues.
