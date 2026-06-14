#!/usr/bin/env bash
# check.sh — run the overpacked-app repo-wide build/test gates.
#
# Mirrors what CI and the team run: `make build` then `make test` from the repo root
# (which themselves invoke the backend Go and frontend vue-tsc/lint checks). Each step is
# best-effort and skipped gracefully if the toolchain isn't available. Exits non-zero if any
# executed gate fails.
#
# Usage: bash check.sh [repo-root]
#   With no argument it locates the repo root by walking up to find the Makefile.

set -uo pipefail

find_root() {
  if [ "${1:-}" != "" ] && [ -f "$1/Makefile" ]; then echo "$1"; return; fi
  dir="$(pwd)"
  while [ "$dir" != "/" ]; do
    if [ -f "$dir/Makefile" ] && [ -d "$dir/backend" ] && [ -d "$dir/frontend" ]; then
      echo "$dir"; return
    fi
    dir="$(dirname "$dir")"
  done
  echo ""
}

ROOT="$(find_root "${1:-}")"
if [ -z "$ROOT" ]; then
  echo "Could not locate the repo root (no Makefile with backend/ and frontend/ found)."
  echo "Pass it explicitly: bash check.sh /path/to/overpacked-app"
  exit 1
fi

cd "$ROOT" || { echo "cannot cd to $ROOT"; exit 1; }
echo "Running repo gates in: $(pwd)"
echo

if ! command -v make >/dev/null 2>&1; then
  echo "make not found on PATH — cannot run the repo gates."
  exit 127
fi

fail=0
run_step() {
  local label="$1"; shift
  echo "==> $label"
  if "$@"; then echo "ok"; else echo "FAILED: $label"; fail=1; fi
  echo
}

run_step "make build" make build
run_step "make test"  make test

if [ "$fail" -eq 0 ]; then
  echo "All executed gates passed."
else
  echo "Some gates failed — see output above."
fi
exit "$fail"
