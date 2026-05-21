#!/usr/bin/env bash
# API curl samples for local verification
# Usage:
#   ./scripts/api-curl-samples.sh              # run full happy-path flow
#   ./scripts/api-curl-samples.sh health       # run a single sample
#   ./scripts/api-curl-samples.sh login
#   ./scripts/api-curl-samples.sh bulk
#
# Requires: curl (jq optional, for token/image_id extraction)

set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
USER_ID="${USER_ID:-4a2bc1d8-7e3f-412e-a19b-625d91c84f32}"
ACCESS_TOKEN="${ACCESS_TOKEN:-}"
REFRESH_TOKEN="${REFRESH_TOKEN:-}"
IMAGE_ID="${IMAGE_ID:-}"

pretty() {
  if command -v jq >/dev/null 2>&1; then
    jq .
  else
    cat
  fi
}

section() {
  echo
  echo "=== $1 ==="
}

# ---------------------------------------------------------------------------
# Individual samples (also callable: ./scripts/api-curl-samples.sh <name>)
# ---------------------------------------------------------------------------

sample_health() {
  section "GET /health"
  curl -sS "${BASE_URL}/health" | pretty
}

sample_login() {
  section "POST /api/v1/auth/login"
  curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"${USER_ID}\"}" | pretty
}

sample_refresh() {
  section "POST /api/v1/auth/refresh"
  if [[ -z "${REFRESH_TOKEN}" ]]; then
    echo "Set REFRESH_TOKEN first (run login flow or export manually)."
    return 1
  fi
  curl -sS -X POST "${BASE_URL}/api/v1/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refresh_token\":\"${REFRESH_TOKEN}\"}" | pretty
}

sample_bulk() {
  section "POST /api/v1/images/bulk"
  if [[ -z "${ACCESS_TOKEN}" ]]; then
    echo "Set ACCESS_TOKEN first (run login flow or export manually)."
    return 1
  fi
  curl -sS -X POST "${BASE_URL}/api/v1/images/bulk" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
      "images": [
        { "original_filename": "landscape.png", "file_type": "image/png" },
        { "original_filename": "portrait.jpg", "file_type": "image/jpeg" }
      ]
    }' | pretty
}

sample_list() {
  section "GET /api/v1/images"
  if [[ -z "${ACCESS_TOKEN}" ]]; then
    echo "Set ACCESS_TOKEN first."
    return 1
  fi
  curl -sS "${BASE_URL}/api/v1/images" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" | pretty
}

sample_get() {
  section "GET /api/v1/images/:id"
  if [[ -z "${ACCESS_TOKEN}" || -z "${IMAGE_ID}" ]]; then
    echo "Set ACCESS_TOKEN and IMAGE_ID first."
    return 1
  fi
  curl -sS "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" | pretty
}

sample_download() {
  section "GET /api/v1/images/:id/download"
  if [[ -z "${ACCESS_TOKEN}" || -z "${IMAGE_ID}" ]]; then
    echo "Set ACCESS_TOKEN and IMAGE_ID first."
    return 1
  fi
  curl -sS "${BASE_URL}/api/v1/images/${IMAGE_ID}/download" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" | pretty
}

sample_update() {
  section "PUT /api/v1/images/:id"
  if [[ -z "${ACCESS_TOKEN}" || -z "${IMAGE_ID}" ]]; then
    echo "Set ACCESS_TOKEN and IMAGE_ID first."
    return 1
  fi
  curl -sS -X PUT "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"original_filename":"updated_v2_landscape.png"}' | pretty
}

sample_delete() {
  section "DELETE /api/v1/images/:id"
  if [[ -z "${ACCESS_TOKEN}" || -z "${IMAGE_ID}" ]]; then
    echo "Set ACCESS_TOKEN and IMAGE_ID first."
    return 1
  fi
  curl -sS -o /dev/null -w "HTTP %{http_code}\n" -X DELETE \
    "${BASE_URL}/api/v1/images/${IMAGE_ID}" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}"
}

# ---------------------------------------------------------------------------
# End-to-end flow
# ---------------------------------------------------------------------------

run_flow() {
  sample_health

  section "Login and capture tokens"
  login_resp="$(curl -sS -X POST "${BASE_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"${USER_ID}\"}")"
  echo "${login_resp}" | pretty

  if command -v jq >/dev/null 2>&1; then
    ACCESS_TOKEN="$(echo "${login_resp}" | jq -r '.access_token')"
    REFRESH_TOKEN="$(echo "${login_resp}" | jq -r '.refresh_token')"
  else
    echo "Install jq to auto-extract tokens, or set ACCESS_TOKEN manually."
    return 0
  fi

  sample_refresh

  section "Bulk upload and capture image_id"
  bulk_resp="$(curl -sS -X POST "${BASE_URL}/api/v1/images/bulk" \
    -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
      "images": [
        { "original_filename": "landscape.png", "file_type": "image/png" }
      ]
    }')"
  echo "${bulk_resp}" | pretty
  IMAGE_ID="$(echo "${bulk_resp}" | jq -r '.records[0].image_id')"

  sample_list
  sample_get
  sample_download
  sample_update
  sample_get
  sample_delete

  echo
  echo "Done. Exported for reuse:"
  echo "  export ACCESS_TOKEN='${ACCESS_TOKEN}'"
  echo "  export REFRESH_TOKEN='${REFRESH_TOKEN}'"
  echo "  export IMAGE_ID='${IMAGE_ID}'"
}

case "${1:-flow}" in
  health)   sample_health ;;
  login)    sample_login ;;
  refresh)  sample_refresh ;;
  bulk)     sample_bulk ;;
  list)     sample_list ;;
  get)      sample_get ;;
  download) sample_download ;;
  update)   sample_update ;;
  delete)   sample_delete ;;
  flow|"")  run_flow ;;
  *)
    echo "Unknown sample: $1"
    echo "Available: health login refresh bulk list get download update delete flow"
    exit 1
    ;;
esac
