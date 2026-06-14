# Testing and Security

## Test organization
- White-box tests live in the same package; use the `_test` package suffix for black-box tests
  that exercise only the public API.
- Test files end in `_test.go` and sit next to the code they test.

## Writing tests
- Use table-driven tests for multiple cases, with `t.Run` subtests so failures are isolated and
  named.
- Name tests descriptively. **In this repo, use `TestXxxHandlerScenario` CamelCase**
  (e.g. `TestItemsHandlerCreateInvalidBody`) to match existing tests — not the
  `Test_functionName_scenario` underscore style. See `references/repo-conventions.md`.
- Cover both success and error paths.
- Reach for `testify` or similar only when it adds real value — don't complicate simple tests
  with assertion libraries.

## Test helpers
- Mark helpers with `t.Helper()` so failures report the caller's line.
- Build fixtures for complex setup.
- Use the `testing.TB` interface for helpers shared between tests and benchmarks.
- Clean up with `t.Cleanup()`.

## Integration tests, the full-stack E2E, and coverage (this repo)
- DB-backed tests gate themselves on `RUN_CONTAINERIZED_TESTS == "true"` + `DATABASE_URL`, run
  migrations once via `sync.Once`, `TRUNCATE … RESTART IDENTITY CASCADE`, then build
  `store.New(db)` and the handler. Pattern source: `internal/http/handlers/labels_test.go`; see
  also `references/repo-conventions.md` for the isolation rules.
- The full-stack E2E (`internal/app/api_e2e_test.go`) drives the **assembled router over real
  HTTP**: `httptest.NewServer(app.NewHTTPHandler(authService, st, appPassword, backupHandler))`
  plus an `http.Client` with a `cookiejar` (so the HttpOnly refresh cookie flows login→refresh).
  This is what covers routing + auth middleware + handlers + stores together — handler-level tests
  call methods directly and bypass the router/middleware. It runs migrations but **not** seeds, so
  empty-list assertions hold. Build the production handler via the exported `app.NewHTTPHandler`,
  not by re-wiring middleware in the test.
- Coverage command (also `make test-backend-container`):
  `go test -p 1 -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...`
  - `-coverpkg=./...` is **required**: without it Go credits only the package under test, so
    `store`/`backup`/`app` code driven by handler/app tests reads as 0%.
  - `-covermode=atomic` for the concurrent paths; `-p 1` because integration packages share one DB.
  - SonarQube ingests `backend/coverage.out`; coverage is scoped to backend app code (see the
    repo-general SonarQube reference).

## Security (this repo's HTTP patterns)
- **Auth tokens:** the refresh token is delivered in an `HttpOnly` / `SameSite=Lax` cookie
  (`op_refresh`, path-scoped to `/api/v1/auth`), never in the JSON body — XSS can't read it. Set
  `Secure` from the request (`r.TLS != nil || X-Forwarded-Proto == "https"`) so it's secure behind
  the prod TLS proxy but still works over local HTTP dev. Do **not** change this conditional to a
  literal `Secure: true` — it breaks HTTP/LAN dev; the `go:S2092` finding on it is intentional and
  accepted in SonarQube.
- **URL inputs:** reject non-`http(s)` schemes on user-supplied URLs before storing them
  (`validateOptionalHTTPURL` in `handlers`), so a `javascript:`/`data:` URL can't be rendered into
  an href later (stored XSS). The frontend mirrors this with `safeHttpUrl` at render sites.
- **Request hardening (middleware in `internal/app`):** `securityHeaders` (nosniff, X-Frame-Options),
  `limitBody` (`http.MaxBytesReader`, exempting the large backup-import upload), plus a per-image
  byte cap in `applyImage`. Untrusted archives (backup import) are read through `io.LimitReader`
  with per-entry **and** aggregate caps to stop decompression bombs.

## Security

### Input validation
- Validate all external input.
- Use strong typing to make invalid states unrepresentable.
- Sanitize data before SQL queries; be careful with file paths derived from user input.
- Escape data for the context it lands in (HTML, SQL, shell).

### Cryptography
- Use the standard library crypto packages; never roll your own crypto.
- Use `crypto/rand` for randomness that must be unpredictable.
- Hash passwords with bcrypt, scrypt, or argon2 (see `golang.org/x/crypto` for options).
- Use TLS for network communication.
