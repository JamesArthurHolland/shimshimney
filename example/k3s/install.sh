#!/usr/bin/env bash
set -euo pipefail

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "This example creates a local k3d cluster on macOS." >&2
  exit 1
fi
if [[ "$(id -u)" -eq 0 ]]; then
  echo "Run this script as your normal user, not with sudo." >&2
  exit 1
fi
if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
  echo "Start Docker Desktop before creating a k3d cluster." >&2
  exit 1
fi
if ! command -v k3d >/dev/null 2>&1; then
  if ! command -v brew >/dev/null 2>&1; then
    echo "Install k3d (for example, with Homebrew) and retry." >&2
    exit 1
  fi
  brew install k3d
fi

cluster="${K3D_CLUSTER_NAME:-shimshimney}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# Host port mapped to Traefik (the k3s default ingress) via the k3d load balancer.
http_port="${HTTP_PORT:-8081}"

if k3d cluster list --no-headers | awk -v name="$cluster" '$1 == name { found = 1 } END { exit !found }'; then
  echo "Starting existing k3d cluster ${cluster} (if stopped)"
  k3d cluster start "$cluster"
  if ! docker port "k3d-${cluster}-serverlb" 80/tcp >/dev/null 2>&1; then
    echo "Mapping host port ${http_port} to the ingress load balancer"
    k3d cluster edit "$cluster" --port-add "${http_port}:80@loadbalancer"
  fi
else
  echo "Creating k3d cluster ${cluster}"
  k3d cluster create "$cluster" --servers 1 --agents 0 \
    --volume "${repo_root}:${repo_root}@server:0" \
    --port "${http_port}:80@loadbalancer" \
    --wait --timeout 120s
fi

kubeconfig="${K3D_KUBECONFIG:-$HOME/.kube/config}"
mkdir -p "$(dirname "$kubeconfig")"
k3d kubeconfig merge "$cluster" --output "$kubeconfig" --kubeconfig-switch-context
echo "Cluster ready: k3d-${cluster}"
echo "Kubeconfig: $kubeconfig"
echo "Verify with: kubectl --kubeconfig \"$kubeconfig\" get nodes"
