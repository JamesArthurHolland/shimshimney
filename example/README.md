# Example projects

`projects/backend-1` through `projects/backend-8` are small Go HTTP servers
using `projects/common`; `GET /hello` returns its service name (for example,
`backend-1`).
`projects/api` requests all eight backends concurrently through their
Kubernetes Services and returns their responses in backend-number order.
It returns HTTP 502 if any backend fails. The shared image in
`projects/dockerfile/Dockerfile` runs the shim, not a hard-coded backend;
`APP_DIR` selects the project (for example,
`/workspace/example/projects/backend-2`). The shim listens on port 9090 and
starts the selected application on port 8080.

## Configuration

`config.yaml` supplies the shim's `operator_url`, `build`, `run`, and
`build_mode` settings. The image uses
`SHIMNEY_CONFIG_PATH=/workspace/example/config.yaml`. Override the file path
with `SHIMNEY_CONFIG_PATH`, or individual values with `SHIMNEY_OPERATOR_URL`,
`SHIMNEY_BUILD`, `SHIMNEY_RUN`, and `SHIMNEY_BUILD_MODE` (also accepts
`SHIMNEY_MODE`). Commands run in `APP_DIR`.

In `hot` mode the shim builds, runs the app, registers with the operator, sends
heartbeats, and exposes `/health`, `/heartbeat`, and `/rebuild` on port 9090.
In `cold` mode it runs the build command and exits without contacting the
operator or starting the app. The example compiles to `/tmp/shimshimney-app`
inside each container so builds do not leave binaries in the shared source.
A hot `/rebuild` currently rebuilds the source but **does not restart** an
already-running application.

## Local image

From the repository root, with Docker available:

```sh
docker build -f example/projects/dockerfile/Dockerfile -t shimshimney/shim:dev .
docker run --rm -p 9090:9090 -p 8080:8080 \
  -e APP_DIR=/workspace/example/projects/backend-1 \
  -e APP_NAME=backend-1 -e POD_NAMESPACE=example shimshimney/shim:dev
```

In another terminal, `curl http://localhost:8080/hello` queries the backend and
`curl http://localhost:9090/health` checks the shim. Without an operator at
the configured URL, registration/heartbeat attempts log warnings.

## Local k3d cluster (macOS)

Start Docker Desktop and create a local k3s cluster with k3d from the
repository root (no `sudo`); see [k3d setup](k3s/README.md):

```sh
./example/k3s/install.sh
kubectl --kubeconfig "$HOME/.kube/config" get nodes
```

You also need Go and `envsubst` for the deployment script:

```sh
./example/scripts/build-and-deploy.sh
```

To start fresh instead, run `./reset-and-deploy.sh` from the repository root.
It deletes the `example` and `shimshimney` namespaces in the selected k3d
cluster, waits for deletion, then runs the same build-and-deploy script to
reinstall the operator and all example projects. This **deletes all resources
in both namespaces**; it does not delete the k3d cluster. Set
`K3D_CLUSTER_NAME` to select a different k3d cluster before running it.

The script runs the operator and its ServiceAccount in the `shimshimney`
namespace, and the API and eight backends in the `example` namespace.
The shim receives its namespace via the pod downward API and registers in
that namespace. The script builds the shim and operator images, imports them into k3d,
compiles the example projects in parallel, installs the operator's CRD and
RBAC, and applies `k8s/namespace.yaml`, `k8s/operator.yaml`, and
`k8s/backend-template.yaml` for each project. The template mounts
`<repository>/example` from the **k3d node's filesystem** at
`/workspace/example`; the setup script bind-mounts the checkout into that
node, so local source edits are visible inside pods. Docker Desktop must
permit sharing this path.
Set `IMAGE_TAG` or `GO_BIN` to override the defaults. Deployments
restart on each run so k3d picks up re-imported images with the same tag.

The API is exposed through Traefik, the ingress controller that k3s ships
by default. `k3s/install.sh` maps host port `HTTP_PORT` (default `8081`) to
the k3d load balancer on port 80. It adds this mapping to existing clusters.
`k8s/api-ingress.yaml` routes `api.localhost` to `service/api`. To query it,
run `./test-scripts/curl-api.sh` from the repository root, or run
`curl http://api.localhost:8081/`. The response is the concatenated `/hello`
output: each backend returns its name followed by `common.RandomString`
(defined in `projects/common/random.go`), for example `backend-1 69f39ccb`.
To change that constant, run `./test-scripts/set-random-string.sh [value]`.
With no value, it generates a random one; add `--rebuild` to also rebuild the
`example` namespace. This is a quick way to see a shared-code change reach
every running backend.

The operator is exposed the same way at `operator.localhost` using
`k8s/operator-ingress.yaml`. Run `./test-scripts/rebuild.sh [namespace]`
(default `example`) to `POST /rebuild`. Each registered shim in that namespace
rebuilds from the mounted source. If the build succeeds, the shim restarts its
app; if it fails, the old app keeps running. The script exits non-zero if any
pod fails. Pods mount the checkout's `example/` and `pkg/` directories from the
k3d node, so edits on the Mac are visible without rebuilding images. Changes to
the shim or operator code itself still require `build-and-deploy.sh`.

To rebuild only registered example pods, use the CLI's
`shimshimney rebuild example` command against the operator API. Rebuild
requests must specify a namespace; the operator will not rebuild pods from
other namespaces. For a local CLI, expose the operator with
`kubectl -n shimshimney port-forward service/operator 8080:8080`.

**Current limitation:** the HTTP registration path currently stores service
definitions in memory rather than creating Kubernetes Services; the
`ShimPod` controller reconciles Services only for `ShimPod` custom resources
created separately. The script deploys explicit backend Services for the
example. Treat these manifests as a local development example.
