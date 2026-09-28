#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cluster="${K3D_CLUSTER_NAME:-shimshimney}"
context="k3d-${cluster}"

for tool in kubectl k3d docker envsubst; do
  if ! command -v "$tool" >/dev/null 2>&1; then
    echo "$tool is required before resetting the cluster" >&2
    exit 1
  fi
done
if ! docker info >/dev/null 2>&1; then
  echo "Start Docker Desktop before resetting the cluster" >&2
  exit 1
fi
if [[ -n "${GO_BIN:-}" ]]; then
  go_bin="$GO_BIN"
elif command -v go1.23.8 >/dev/null 2>&1; then
  go_bin=go1.23.8
else
  go_bin=go
fi
if ! command -v "$go_bin" >/dev/null 2>&1; then
  echo "Go binary $go_bin is required before resetting the cluster" >&2
  exit 1
fi
kubectl --context "$context" get nodes >/dev/null

echo "==> deleting example and shimshimney namespaces from $context"
kubectl --context "$context" delete namespace example shimshimney \
  --ignore-not-found --wait=true --timeout=300s

echo "==> rebuilding and deploying operator and example"
exec "$repo_root/example/scripts/build-and-deploy.sh"
