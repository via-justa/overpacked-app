#!/usr/bin/env bash
set -euo pipefail

COMPOSE_CMD="${COMPOSE_CMD:-docker compose}"
API_BASE_URL="${API_BASE_URL:-http://localhost:8000}"
KEEP_STACK="${KEEP_STACK:-false}"
WAIT_TIMEOUT_SECONDS="${WAIT_TIMEOUT_SECONDS:-120}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROOT_ENV_FILE="${ROOT_DIR}/.env"
TEMP_ENV_CREATED="false"

if ! command -v docker >/dev/null 2>&1; then
  echo "error: docker is required" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "error: curl is required" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq is required" >&2
  exit 1
fi

if [[ ! -x "${ROOT_DIR}/dev/scripts/full_api_curl_test.sh" ]]; then
  echo "error: dev/scripts/full_api_curl_test.sh must exist and be executable" >&2
  exit 1
fi

cleanup() {
  if [[ "${KEEP_STACK}" != "true" ]]; then
    echo ""
    echo "==> stopping compose stack"
    ${COMPOSE_CMD} -f "${ROOT_DIR}/docker-compose.yml" down --volumes
  fi

  if [[ "${TEMP_ENV_CREATED}" == "true" ]]; then
    rm -f "${ROOT_ENV_FILE}"
  fi
}
trap cleanup EXIT

if [[ ! -f "${ROOT_ENV_FILE}" ]]; then
  echo "==> .env not found, creating temporary defaults for compose run"
  cat >"${ROOT_ENV_FILE}" <<EOF
DATABASE_URL=postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@db:5432/${POSTGRES_DB:-packing_light}?sslmode=disable
SERVER_ADDR=0.0.0.0:8000
APP_USERNAME=${APP_USERNAME:-admin}
APP_PASSWORD=${APP_PASSWORD:-pw123}
JWT_SECRET=${JWT_SECRET:-test-secret}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
POSTGRES_DB=${POSTGRES_DB:-packing_light}
EOF
  TEMP_ENV_CREATED="true"
fi

echo "==> starting db and backend services"
${COMPOSE_CMD} -f "${ROOT_DIR}/docker-compose.yml" up -d db backend

echo "==> waiting for API health at ${API_BASE_URL}/health"
start_time="$(date +%s)"
while true; do
  if curl -sS "${API_BASE_URL}/health" >/dev/null 2>&1; then
    if [[ "$(curl -sS -o /dev/null -w '%{http_code}' "${API_BASE_URL}/health")" == "200" ]]; then
      break
    fi
  fi

  now="$(date +%s)"
  elapsed="$((now - start_time))"
  if (( elapsed >= WAIT_TIMEOUT_SECONDS )); then
    echo "error: API did not become healthy within ${WAIT_TIMEOUT_SECONDS}s" >&2
    ${COMPOSE_CMD} -f "${ROOT_DIR}/docker-compose.yml" logs backend || true
    exit 1
  fi

  sleep 2
done

echo "==> running full API curl test"
APP_USERNAME="${APP_USERNAME:-admin}" \
APP_PASSWORD="${APP_PASSWORD:-pw123}" \
API_BASE_URL="${API_BASE_URL}" \
"${ROOT_DIR}/dev/scripts/full_api_curl_test.sh"

echo ""
echo "compose-backed full API curl test completed successfully"
