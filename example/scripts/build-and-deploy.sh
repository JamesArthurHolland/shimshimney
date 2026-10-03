#!/usr/bin/env bash
# Builds the shim runtime image and deploys every example project to the
# local k3d cluster in parallel, then applies the operator and one Deployment
# per project using the shared shim/hostPath-volume template.
#
# Requires: docker, k3d, kubectl, envsubst, and Go >= 1.22.
#
# Usage:
#   ./example/scripts/build-and-deploy.sh
#
# Env vars:
#   IMAGE_TAG   image tag to build/deploy for the shim and operator (default: dev)
#   GO_BIN      Go binary used to build the projects (default: go1.23.8 if available)
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

image_tag="${IMAGE_TAG:-dev}"
namespace="example"
operator_namespace="shimshimney"

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required" >&2; exit 1; }
command -v k3d >/dev/null 2>&1 || { echo "k3d is required; run ./example/k3s/install.sh" >&2; exit 1; }
command -v envsubst >/dev/null 2>&1 || { echo "envsubst is required" >&2; exit 1; }
cluster="${K3D_CLUSTER_NAME:-shimshimney}"
kubectl --context "k3d-${cluster}" get nodes >/dev/null

echo "==> building shim runtime image (shimshimney/shim-example-go:${image_tag})"
docker build -f example/projects/dockerfile/Dockerfile -t "shimshimney/shim-example-go:${image_tag}" .

echo "==> building operator image (shimshimney/operator:${image_tag})"
docker build -f operator/Dockerfile -t "shimshimney/operator:${image_tag}" .
k3d image import "shimshimney/shim-example-go:${image_tag}" "shimshimney/operator:${image_tag}" -c "$cluster"

# discover every project directory under example/projects except the
# infrastructure helpers (common, dockerfile)
projects=()
for dir in example/projects/*/; do
  name="$(basename "$dir")"
  case "$name" in
    common|dockerfile) continue ;;
  esac
  projects+=("$name")
done

echo "==> found ${#projects[@]} project(s): ${projects[*]}"

if [[ -n "${GO_BIN:-}" ]]; then
  go_bin="$GO_BIN"
elif command -v go1.23.8 >/dev/null 2>&1; then
  go_bin="go1.23.8"
else
  go_bin="go"
fi
command -v "$go_bin" >/dev/null 2>&1 || { echo "Go binary $go_bin not found" >&2; exit 1; }

echo "==> building each project in parallel"
build_pids=()
for name in "${projects[@]}"; do
  (
    echo "--> building ${name}"
    (cd "example/projects/${name}" && "$go_bin" build -o /dev/null .)
  ) &
  build_pids+=($!)
done

build_status=0
for pid in "${build_pids[@]}"; do
  wait "$pid" || build_status=1
done

if [ "$build_status" -ne 0 ]; then
  echo "one or more project builds failed" >&2
  exit "$build_status"
fi

echo "==> applying namespaces and operator"
export NAMESPACE="$namespace" OPERATOR_NAMESPACE="$operator_namespace" IMAGE_TAG="$image_tag"
envsubst '${NAMESPACE}' < example/k8s/namespace.yaml | kubectl --context "k3d-${cluster}" apply -f -
NAMESPACE="$operator_namespace" envsubst '${NAMESPACE}' < example/k8s/namespace.yaml | kubectl --context "k3d-${cluster}" apply -f -
kubectl --context "k3d-${cluster}" apply -f operator/config/crd/bases/shimshimney.shimshimney.io_shimpods.yaml
kubectl --context "k3d-${cluster}" apply -f operator/config/rbac/role.yaml
envsubst '${OPERATOR_NAMESPACE}' < example/k8s/operator-rbac.yaml | kubectl --context "k3d-${cluster}" apply -f -
envsubst '${OPERATOR_NAMESPACE} ${IMAGE_TAG}' < example/k8s/operator.yaml | kubectl --context "k3d-${cluster}" apply -f -
envsubst '${OPERATOR_NAMESPACE}' < example/k8s/operator-ingress.yaml | kubectl --context "k3d-${cluster}" apply -f -
# Remove only the example resources created by earlier versions of this script.
for name in "${projects[@]}"; do
  kubectl --context "k3d-${cluster}" delete "deployment/${name}" "service/${name}" -n "$operator_namespace" --ignore-not-found
done
kubectl --context "k3d-${cluster}" rollout restart deployment/operator -n "$operator_namespace"

deploy_project() {
  local name="$1"
  NAME="$name" REPO_PATH="$repo_root" envsubst '${NAME} ${REPO_PATH} ${NAMESPACE} ${OPERATOR_NAMESPACE} ${IMAGE_TAG}' \
    < example/k8s/backend-template.yaml \
    | kubectl --context "k3d-${cluster}" apply -n "$namespace" -f -
}

pids=()
for name in "${projects[@]}"; do
  deploy_project "$name" &
  pids+=($!)
done

status=0
for pid in "${pids[@]}"; do
  wait "$pid" || status=1
done

if [ "$status" -ne 0 ]; then
  echo "one or more deployments failed" >&2
  exit "$status"
fi
envsubst '${NAMESPACE}' < example/k8s/api-ingress.yaml | kubectl --context "k3d-${cluster}" apply -f -

echo "==> waiting for rollouts"
kubectl --context "k3d-${cluster}" rollout status "deployment/operator" -n "$operator_namespace" --timeout=120s
pids=()
for name in "${projects[@]}"; do
  kubectl --context "k3d-${cluster}" rollout restart "deployment/${name}" -n "$namespace"
  kubectl --context "k3d-${cluster}" rollout status "deployment/${name}" -n "$namespace" --timeout=600s &
  pids+=($!)
done
status=0
for pid in "${pids[@]}"; do
  wait "$pid" || status=1
done
if [ "$status" -ne 0 ]; then
  echo "one or more project rollouts failed" >&2
  exit "$status"
fi

echo "==> done"
