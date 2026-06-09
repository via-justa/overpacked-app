# Types, Interfaces, and Project Structure

## Type definitions
- Define named types to add meaning and type safety, not just aliases.
- Use struct tags for JSON/XML/database mappings.
- Prefer explicit type conversions.
- When using type assertions, check the second (`, ok`) return value.
- Prefer generics with constraints over unconstrained types. When you truly need an
  unconstrained type, use the predeclared `any` (Go 1.18+), not `interface{}`.

## Pointers vs values
The choice is about mutation, size, and the zero value — not habit. Be consistent within a
single type's method set.

- Pointer receivers: large structs, or when the method must modify the receiver.
- Value receivers: small structs, or when you want immutability.
- Pointer parameters: when you must modify the argument, or for large structs.
- Value parameters: small structs, or to guarantee the callee can't mutate the caller's data.
- Consider how the zero value behaves when deciding — a useful zero value often favors value
  semantics.

## Interfaces and composition
- Accept interfaces, return concrete types. This keeps callers flexible while giving them the
  full concrete API on what they receive.
- Keep interfaces small (1–3 methods).
- Use embedding for composition rather than deep type hierarchies.
- Define interfaces close to where they're *used*, not where they're implemented — the
  consumer knows what it needs.
- Don't export interfaces unless callers genuinely need them.

## Project structure
- Follow the standard Go project layout.
- `main` packages live under `cmd/`.
- Reusable packages go in `pkg/` or `internal/`; use `internal/` for anything that must not be
  imported by external projects.
- Group related functionality into packages and avoid circular dependencies.

## Dependency management
- Use Go modules (`go.mod` / `go.sum`); keep dependencies minimal.
- Run `go mod tidy` to drop unused dependencies; update regularly for security patches.
- Vendor only when there's a concrete reason to.
