# Contributing

Thanks for your interest in contributing.

## Code of Conduct

Please be respectful and constructive in all interactions.

## Development Setup

### Prerequisites

- Go (matching `backend/go.mod`)
- Node.js and npm (matching `frontend/package.json`)
- Docker (optional but recommended)

### Install dependencies

```bash
make install
```

### Run locally

```bash
make up
```

Or run parts separately:

```bash
make backend
make frontend
```

## Common Commands

```bash
make test
make test-backend
make test-frontend
make build
```

## Contribution Workflow

1. Create an issue first for significant changes.
2. Fork the repository to your GitHub account.
3. Clone your fork and add the main repository as `upstream`.
4. Create a branch from your fork's up-to-date `main`.
5. Keep PRs focused and small.
6. Run relevant tests before opening a PR.
7. Fill in the PR template completely.

### Fork-based workflow (required)

```bash
# Clone your fork
git clone git@github.com:<your-user>/overpacked-app.git
cd overpacked-app

# Add main repo as upstream
git remote add upstream git@github.com:via-justa/overpacked-app.git

# Sync main
git fetch upstream
git checkout main
git merge --ff-only upstream/main

# Create your feature branch
git checkout -b feat/my-change
```

## Project-Specific Rules

### API changes (spec-first)

When changing backend API behavior:

1. Update `dev/openapi.yaml` first.
2. Regenerate API code:

```bash
make gen-api-go
```

3. Implement/update handlers.

### Database changes

- Add a new migration file under `backend/internal/migrations/sql`.
- Include both up and down migration blocks.
- Do not edit historical migrations that may already be applied in shared environments.
- Sync docs if schema changes:
  - `dev/database-schema.mermaid`
  - `dev/openapi.yaml` (if API-impacting)

### Frontend notes

- Use semantic design tokens from `frontend/src/style.css`.
- Avoid raw Tailwind palette classes in app/feature templates.
- Run theme lint for UI changes:

```bash
cd frontend && npm run lint:theme
```

## Commit and PR Guidance

- Use clear commit messages.
- Use PR titles with release keywords when possible (for example: `(feat)`, `(fix)`, `(chore)`).
- Link issues in the PR description.
- Include screenshots for UI changes.

## Reporting Bugs / Requesting Features

Use the issue templates:

- Bug report
- Feature request

## Security

Do not commit secrets, tokens, credentials, or private keys.
If you discover a security issue, please report it privately to maintainers.
