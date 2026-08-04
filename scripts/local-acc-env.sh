#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# Export environment variables for provider acceptance tests against a local Graylog.
#
# Usage (from repo root):
#   source scripts/local-acc-env.sh
#   # optional overrides:
#   GRAYLOG_ENDPOINT=http://localhost:9000/api GRAYLOG_PASSWORD=... source scripts/local-acc-env.sh
#
# Then run:
#   TF_ACC=1 make testacc
#   # or
#   TF_ACC=1 go test -v -count=1 ./internal/provider/ -timeout 120m

set -euo pipefail

GRAYLOG_ENDPOINT="${GRAYLOG_ENDPOINT:-http://localhost:9000/api}"
GRAYLOG_USERNAME="${GRAYLOG_USERNAME:-admin}"
GRAYLOG_PASSWORD="${GRAYLOG_PASSWORD:-}"

if [[ -z "${GRAYLOG_PASSWORD}" ]]; then
  # Default matches the plaintext used for GRAYLOG_ROOT_PASSWORD_SHA2 in
  # a typical local dev-workspace/.env (same value as GRAYLOG_PASSWORD_SECRET).
  if [[ -f "dev-workspace/.env" ]]; then
    # shellcheck disable=SC1091
    source "dev-workspace/.env"
  fi
  if [[ -f "dev-workspace/graylog-demo/terraform.tfvars" ]]; then
    # Prefer the password already configured for the local demo stack.
    demo_pw="$(sed -n 's/^graylog_password[[:space:]]*=[[:space:]]*"\(.*\)"/\1/p' dev-workspace/graylog-demo/terraform.tfvars | head -1)"
    if [[ -n "${demo_pw}" ]]; then
      GRAYLOG_PASSWORD="${demo_pw}"
    fi
  fi
fi

if [[ -z "${GRAYLOG_PASSWORD}" ]]; then
  echo "GRAYLOG_PASSWORD is required (set it, or configure dev-workspace/graylog-demo/terraform.tfvars)" >&2
  return 1 2>/dev/null || exit 1
fi

export GRAYLOG_ENDPOINT GRAYLOG_USERNAME GRAYLOG_PASSWORD
export TF_ACC="${TF_ACC:-1}"

# Discover default index set ID when Graylog is reachable.
if [[ -z "${GRAYLOG_DEFAULT_INDEX_SET_ID:-}" ]]; then
  if command -v curl >/dev/null 2>&1; then
    idx_json="$(curl -sf -u "${GRAYLOG_USERNAME}:${GRAYLOG_PASSWORD}" \
      -H 'Accept: application/json' \
      -H 'X-Requested-By: terraform-provider-graylog' \
      "${GRAYLOG_ENDPOINT}/system/indices/index_sets?skip=0&limit=1" 2>/dev/null || true)"
    if [[ -n "${idx_json}" ]]; then
      if command -v jq >/dev/null 2>&1; then
        GRAYLOG_DEFAULT_INDEX_SET_ID="$(printf '%s' "${idx_json}" | jq -r '.index_sets[0].id // empty')"
      else
        GRAYLOG_DEFAULT_INDEX_SET_ID="$(printf '%s' "${idx_json}" | sed -n 's/.*"id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)"
      fi
    fi
  fi
fi

if [[ -n "${GRAYLOG_DEFAULT_INDEX_SET_ID:-}" ]]; then
  export GRAYLOG_DEFAULT_INDEX_SET_ID
fi

cat <<EOF
# Local Graylog acceptance-test environment
export GRAYLOG_ENDPOINT=$(printf '%q' "${GRAYLOG_ENDPOINT}")
export GRAYLOG_USERNAME=$(printf '%q' "${GRAYLOG_USERNAME}")
export GRAYLOG_PASSWORD=$(printf '%q' "${GRAYLOG_PASSWORD}")
export TF_ACC=$(printf '%q' "${TF_ACC}")
EOF

if [[ -n "${GRAYLOG_DEFAULT_INDEX_SET_ID:-}" ]]; then
  echo "export GRAYLOG_DEFAULT_INDEX_SET_ID=$(printf '%q' "${GRAYLOG_DEFAULT_INDEX_SET_ID}")"
else
  echo "# GRAYLOG_DEFAULT_INDEX_SET_ID not discovered yet (is Graylog up and past preflight?)" >&2
fi
