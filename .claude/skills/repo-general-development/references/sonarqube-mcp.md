# SonarQube MCP Server

When a SonarQube MCP server is available, follow these guidelines during a task.

## Basic workflow

- **At the start of a task:** if a `toggle_automatic_analysis` tool exists, **disable** automatic
  analysis.
- **At the very end of a task,** after you've finished generating or modifying code files: call
  `analyze_file_list` (if it exists) to analyze the files you created or changed.
- **Then re-enable** automatic analysis with `toggle_automatic_analysis` (if it exists).

## Project keys

- When the user mentions a project key, call `search_my_sonarqube_projects` first to find the
  exact key. Don't guess keys.

## Language detection

- Detect the language from the code's syntax. If unclear, ask or make an educated guess.

## Branch / PR context

- Many operations support branch-specific analysis — if the user is on a feature branch, pass the
  branch parameter.

## Issues and violations

- After fixing issues, **don't** verify them via `search_sonar_issues_in_projects` — the server
  won't reflect the updates yet.

## Troubleshooting

- **Auth:** SonarQube requires **USER** tokens (not project tokens). A
  `SonarQube answered with Not authorized` error usually means the wrong token type.
- **Project not found:** use `search_my_sonarqube_projects`; verify key spelling/format.
- **Analysis quality:** specify the language correctly, provide full file content (snippet
  analysis is weaker), and remember snippet analysis doesn't replace a full project scan.

## This project's setup (overpacked-app)

Analysis runs in **CI**, not via IDE/automatic analysis: `.github/workflows/sonar.yml` runs
`SonarSource/sonarqube-scan-action` against SonarQube Cloud, reading `sonar-project.properties`.

### `sonar-project.properties` conventions
- `sonar.exclusions` — generated + non-source files: `frontend/src/lib/api/schema.ts`,
  `backend/internal/api/api.gen.go`, `dev/test-data.sql`, `.claude/**`.
- `sonar.cpd.exclusions=**/*_test.go` — table-driven test setup is legitimately repetitive.
- Coverage: `sonar.go.coverage.reportPaths=backend/coverage.out` (produced by the CI test step /
  `make test-backend-container`). **Scope** coverage with
  `sonar.coverage.exclusions=frontend/**,backend/cmd/**,…,**/*_test.go` — importing a Go coverage
  report otherwise makes Sonar grade the test-free frontend's new code as 0% and fail the 80%
  new-code gate.

### CI gotchas
- The TS analyzer needs `node_modules` present so `tsconfig`'s `extends: @vue/tsconfig/...`
  resolves — the workflow runs `npm ci --ignore-scripts`. Without it Sonar logs
  "referenced/extended tsconfig.json was not found" and degrades TS analysis.
- A Postgres service + `RUN_CONTAINERIZED_TESTS=true` + `DATABASE_URL` make the integration tests
  run and emit coverage in CI; the test step must run **before** the scan step.
- Pin GitHub Actions to commit SHAs.

### Triaging findings via the MCP
Use `change_sonar_issue_status` (`accept`/`falsepositive`) and `change_security_hotspot_status`
(`REVIEWED` + `SAFE`, with a comment) for findings that are intentional or false positives — don't
contort correct code to satisfy a rule. Examples accepted in this repo: the conditional cookie
`Secure` flag (`go:S2092`), the scheduler's stored base context (`godre:S8242`/`S8239`),
server-controlled dynamic SQL in `store/search.go` (`go:S2077`), and cosmetic `Math.random` for
label colors (`typescript:S2245`). Generated code and dev seed data are excluded rather than
"fixed".
