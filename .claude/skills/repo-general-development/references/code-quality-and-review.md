# Code Quality & Review Standards

Repo-wide quality bar and how to review changes. Language-specific style lives in the
`go-development` and `vue-development` skills; this is the cross-cutting layer.

## Quality standards (apply everywhere)

- **Descriptive, meaningful names** for variables, functions, types.
- **Single responsibility** — each function/type does one thing; keep functions small and
  focused.
- **DRY** — no needless duplication; extract shared logic.
- **Avoid magic numbers and strings** — use named constants, especially for repeated values
  (error messages, routes, query keys, header names, status labels).
- **Avoid deep nesting** (max ~3–4 levels); prefer guard clauses / early returns.
- **Cognitive complexity:** keep functions below **15** (the Sonar threshold). When a function
  exceeds it, extract helpers, simplify conditionals, or split the logic.
- **Self-documenting code first**; comment only non-obvious logic, and keep comments to 1–2 lines
  (complex computed values, keyboard/navigation handlers, position math, state patterns). Don't
  restate obvious code.
- **Proper error handling** — meaningful messages, no silent failures, validate inputs early.

## Review priorities

Prioritize review findings in this order:

**🔴 CRITICAL (block merge)** — security (vulns, exposed secrets, auth/authz), correctness
(logic errors, data-corruption/race risks), breaking API-contract changes without versioning,
data-loss risk.

**🟡 IMPORTANT (needs discussion)** — severe code-quality/SOLID violations or heavy duplication,
missing tests for critical paths or new functionality, obvious performance problems (N+1
queries, leaks), significant deviations from established patterns.

**🟢 SUGGESTION (non-blocking)** — readability/naming, non-functional optimizations, minor
convention deviations, missing docs/comments.

## Review principles

Be specific (cite exact file/line), explain *why* it matters and the impact, suggest a concrete
fix, stay constructive, acknowledge good work, be pragmatic, and group related comments instead
of scattering many about one topic.

### Comment format

```markdown
**[PRIORITY] Category: Brief title**

Detailed description of the issue or suggestion.

**Why this matters:** impact / reasoning.

**Suggested fix:** [code example if applicable]
```

## Repo-specific review checklist

Beyond the general bar, verify for this repo:

- PR title follows the `(type) description` / release-keyword pattern.
- Database migrations are **reversible** (have a `Down` block) and don't edit applied migrations.
- API changes are based on an updated `dev/openapi.yaml` (spec-first), with code regenerated —
  not hand-edited generated files.
- Schema changes are reflected in `dev/database-schema.mermaid`.
- Frontend code uses **semantic design tokens** (no raw Tailwind palette classes) and the icon
  registry / `AppIcon` (no direct PrimeIcons). (The frontend lint scripts enforce both.)
- Weights/volumes use canonical units (grams / ml).
- Dockerfiles follow best practices (multi-stage, minimal base images) and contain **no secrets**.
- Docker compose / Helm charts are updated when services are added, removed, or changed.
- No secrets, tokens, or PII in code or logs.

## Security review focus

Parameterized SQL (never string-concatenated queries), validated/sanitized inputs, auth checks
before resource access, established crypto libraries only, and dependency vulnerability awareness.
