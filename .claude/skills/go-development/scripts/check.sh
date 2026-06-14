#!/usr/bin/env bash
# check.sh — run the standard Go quality gate over a path (default ./...).
#
# Each step is best-effort: if a tool isn't installed or there are no Go files,
# that step is skipped rather than failing the whole run. The script reports only
# actionable output and exits non-zero if any executed check failed.
#
# Usage: bash check.sh [path]   # path defaults to ./...

set -uo pipefail

# This repo's Go module lives in backend/. When run from the repo root for the
# whole module (the default), switch into it so `go` finds the module.
if [ "${1:-./...}" = "./..." ] && [ ! -f go.mod ] && [ -f backend/go.mod ]; then
  cd backend || exit 1
fi

TARGET="${1:-./...}"
# For gofmt we need a directory/file, not a package pattern.
FMT_PATH="."
if [ "$TARGET" != "./..." ] && [ -e "$TARGET" ]; then
  FMT_PATH="$TARGET"
fi

fail=0

have() { command -v "$1" >/dev/null 2>&1; }

if ! have go; then
  echo "go not found on PATH — cannot run checks."
  exit 127
fi

echo "==> gofmt (formatting)"
unformatted="$(gofmt -l "$FMT_PATH" 2>/dev/null)"
if [ -n "$unformatted" ]; then
  echo "These files are not gofmt-clean (run: gofmt -w):"
  echo "$unformatted"
  fail=1
else
  echo "ok"
fi

echo "==> go vet (suspicious constructs)"
if ! go vet "$TARGET"; then
  fail=1
fi

echo "==> golangci-lint"
if have golangci-lint; then
  if ! golangci-lint run "$TARGET"; then
    fail=1
  fi
else
  echo "golangci-lint not installed — skipping (install: https://golangci-lint.run)"
fi

echo "==> go test"
if ! go test "$TARGET"; then
  fail=1
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "All executed checks passed."
else
  echo "Some checks failed — see output above."
fi
exit "$fail"
