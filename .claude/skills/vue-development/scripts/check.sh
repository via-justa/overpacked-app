#!/usr/bin/env bash
# check.sh — run the overpacked-app frontend quality gates.
#
# Mirrors the checks CI applies: TypeScript type-check (vue-tsc), stylelint, and the two
# custom lint scripts (theme palette + direct-primeicons). Each step is best-effort and
# skipped gracefully if the tool or script isn't present. Exits non-zero if any executed
# check failed.
#
# Usage: bash check.sh [frontend-dir]
#   With no argument it locates the frontend directory automatically.

set -uo pipefail

find_frontend() {
  # 1) explicit arg
  if [ "${1:-}" != "" ] && [ -f "$1/package.json" ]; then echo "$1"; return; fi
  # 2) cwd is the frontend
  if [ -f "package.json" ] && grep -q '"vue"' package.json 2>/dev/null; then echo "."; return; fi
  # 3) a frontend/ under cwd
  if [ -f "frontend/package.json" ]; then echo "frontend"; return; fi
  # 4) walk up looking for frontend/package.json (skill may live in .claude/skills/...)
  dir="$(pwd)"
  while [ "$dir" != "/" ]; do
    if [ -f "$dir/frontend/package.json" ]; then echo "$dir/frontend"; return; fi
    dir="$(dirname "$dir")"
  done
  echo ""
}

FE="$(find_frontend "${1:-}")"
if [ -z "$FE" ]; then
  echo "Could not locate the frontend directory (no frontend/package.json found)."
  echo "Pass it explicitly: bash check.sh path/to/frontend"
  exit 1
fi

cd "$FE" || { echo "cannot cd to $FE"; exit 1; }
echo "Running frontend checks in: $(pwd)"
echo

if ! command -v npm >/dev/null 2>&1; then
  echo "npm not found on PATH — cannot run checks."
  exit 127
fi

if [ ! -d node_modules ]; then
  echo "node_modules missing — run 'npm install' in $FE first. Skipping checks."
  exit 1
fi

fail=0
run_step() {
  local label="$1"; shift
  echo "==> $label"
  if "$@"; then echo "ok"; else echo "FAILED: $label"; fail=1; fi
  echo
}

# Type check (vue-tsc). -b uses the project references in tsconfig.
run_step "vue-tsc (type check)" npx --no-install vue-tsc -b --pretty false

# stylelint
if npm run | grep -q "lint:css"; then
  run_step "stylelint" npm run --silent lint:css
fi

# custom: no raw palette classes
if npm run | grep -q "lint:theme"; then
  run_step "theme palette lint" npm run --silent lint:theme
fi

# custom: no direct primeicons
if npm run | grep -q "lint:icons"; then
  run_step "primeicons lint" npm run --silent lint:icons
fi

if [ "$fail" -eq 0 ]; then
  echo "All executed checks passed."
else
  echo "Some checks failed — see output above."
fi
exit "$fail"
