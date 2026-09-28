#!/usr/bin/env bash
# Curls the example API through the k3d load balancer and Traefik ingress,
# printing the concatenated backend responses.
set -euo pipefail

cluster="${K3D_CLUSTER_NAME:-shimshimney}"
host="${API_HOST:-api.localhost}"

command -v curl >/dev/null 2>&1 || { echo "curl is required" >&2; exit 1; }

if [ -z "${HTTP_PORT:-}" ]; then
  mapping="$(docker port "k3d-${cluster}-serverlb" 80/tcp 2>/dev/null | head -n1 || true)"
  if [ -z "$mapping" ]; then
    echo "k3d cluster ${cluster} has no ingress port mapping; run example/k3s/install.sh" >&2
    exit 1
  fi
  HTTP_PORT="${mapping##*:}"
fi

curl -fsS --resolve "${host}:${HTTP_PORT}:127.0.0.1" "http://${host}:${HTTP_PORT}/"
