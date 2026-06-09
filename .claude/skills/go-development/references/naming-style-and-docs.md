# Naming, Style, Formatting, and Documentation

## Naming conventions

### Packages
- Lowercase, single word. No underscores, hyphens, or mixedCaps.
- Name by what the package *provides*, not what it contains. Avoid generic catch-alls like
  `util`, `common`, or `base`.
- Singular, not plural.

### Variables and functions
- Use mixedCaps / MixedCaps (camelCase), never underscores.
- Short but descriptive. Single-letter names only in very short scopes (loop indices, etc.).
- Exported names start uppercase; unexported start lowercase.
- Avoid stuttering: prefer `http.Server` over `http.HTTPServer`. The package name is already
  context, so don't repeat it in the identifier.

### Interfaces
- Name single-method interfaces after the method plus `-er`: `Read` → `Reader`, `Write` →
  `Writer`, `Format` → `Formatter`.
- Keep interfaces small and focused (1–3 methods is ideal).

### Constants
- MixedCaps for exported, mixedCaps for unexported.
- Group related constants in a `const` block.
- Consider typed constants for type safety.

## Formatting
- Always format with `gofmt`; manage imports with `goimports`.
- No hard line-length limit — optimize for readability.
- Use blank lines to separate logical groups within a function.

## Comments
- Prefer self-documenting code; clear names and structure beat narration.
- Comment only to explain complex logic, business rules, or non-obvious behavior. Document
  *why*, not *what*, unless the *what* is genuinely complex.
- Write complete sentences in English (translate only on request).
- Start a comment with the name of the thing it describes. Package comments start with
  "Package [name]".
- Use line comments (`//`) for most things; reserve block comments (`/* */`) mainly for
  package documentation.
- No emoji.

## Documentation
- Document every exported symbol (types, functions, methods, packages) with a concise
  explanation that starts with the symbol name.
- Add runnable examples (`Example` functions) where they clarify usage.
- Keep docs next to the code and update them when the code changes.

### README / project docs
- Clear setup instructions; documented dependencies and requirements.
- Usage examples and configuration options.
- A troubleshooting section.

## Tooling and workflow
Essential tools: `go fmt` (format), `go vet` (suspicious constructs), `golangci-lint`
(linting — `golint` is deprecated), `go test`, `go mod`, `go generate`.

Practices: run tests before committing, use pre-commit hooks for fmt/lint, keep commits
focused and atomic with meaningful messages, and review diffs before committing.
