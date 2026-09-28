#!/usr/bin/env bash
# Asks the operator (via the k3d load balancer and Traefik ingress) to rebuild
# and restart every registered shim in a namespace. Usage: rebuild.sh [namespace]
set -euo pipefail

namespace="${1:-${NAMESPACE:-example}}"
cluster="${K3D_CLUSTER_NAME:-shimshimney}"
host="${OPERATOR_HOST:-operator.localhost}"

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }

if [ -z "${HTTP_PORT:-}" ]; then
  mapping="$(docker port "k3d-${cluster}-serverlb" 80/tcp 2>/dev/null | head -n1 || true)"
  if [ -z "$mapping" ]; then
    echo "k3d cluster ${cluster} has no ingress port mapping; run example/k3s/install.sh" >&2
    exit 1
  fi
  HTTP_PORT="${mapping##*:}"
fi

echo "==> rebuilding namespace ${namespace}" >&2
response="$(curl -fsS --max-time "${REBUILD_TIMEOUT:-300}" \
  --resolve "${host}:${HTTP_PORT}:127.0.0.1" \
  -H 'Content-Type: application/json' \
  -d "{\"namespace\":\"${namespace}\"}" \
  "http://${host}:${HTTP_PORT}/rebuild")"

if command -v jq >/dev/null 2>&1; then
  echo "$response" | jq -r '.results[]'
  echo "$response" | jq -e 'all(.results[]; endswith(" rebuilt"))' >/dev/null
else
  echo "$response"
  ! grep -q ' failed' <<<"$response"
fi
