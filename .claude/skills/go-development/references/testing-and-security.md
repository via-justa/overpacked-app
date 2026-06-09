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
