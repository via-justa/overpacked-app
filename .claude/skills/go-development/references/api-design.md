# API Design (backend)

The backend's HTTP layer is **chi + oapi-codegen**: handlers implement the generated
`ServerInterface`, are aggregated by `apiServer`, and are mounted with `api.HandlerWithOptions`
onto a `chi.NewRouter()` that applies `chimiddleware` (Recoverer, RequestID, Logger) — see
`internal/app/app.go` and `routes.go`. Handler structure, the `writeJSON` / `decodeJSON` helpers,
and the domain-error → HTTP-status mapping are documented in `references/repo-conventions.md`;
follow that for any handler or routing work. This file covers the two adjacent concerns it
doesn't: JSON shaping and outbound HTTP clients.

## JSON APIs

- Control marshaling with struct tags.
- Validate input data.
- Use pointers for optional fields to distinguish "absent" from "zero" — the codebase uses
  `*string` / `*float64` for nullable fields throughout (converted via `sql_helpers.go`).
- Consider `json.RawMessage` to defer parsing of a sub-document.
- Handle JSON encode/decode errors explicitly (handlers do this via `decodeJSON`).

## Outbound HTTP clients

The backend is mostly a server, but it does make outbound calls — e.g. `routePreviewClient` in
`internal/http/handlers/trips.go`, which fetches route previews. When adding an outbound client,
follow that pattern:

- **Configure the `*http.Client` once** (timeouts, transport) — `routePreviewClient` is a
  package-level client. It's safe for concurrent use; don't mutate `Transport` after first use.
- **Build a fresh request per call** with `http.NewRequestWithContext(ctx, ...)` and thread
  `context.Context` through so calls are cancellable.
- **Never stash per-request state** (a `*http.Request`, URL params, body, headers) on a
  long-lived client struct — that corrupts requests under concurrency.
- **Always `defer resp.Body.Close()`** and handle the close path.
