# Local k3d cluster (macOS)

Start Docker Desktop, then from the repository root run **without sudo**:

```sh
./example/k3s/install.sh
kubectl --kubeconfig "$HOME/.kube/config" get nodes
```

The script installs `k3d` with Homebrew if it is missing, creates a
single-server k3s cluster named `shimshimney` in Docker, and selects its
`k3d-shimshimney` context in `~/.kube/config`. If the cluster already exists,
it starts it instead of replacing it. Set `K3D_CLUSTER_NAME` to use another
name or `K3D_KUBECONFIG` to choose another config file. If your `KUBECONFIG`
contains multiple paths, use the explicit `--kubeconfig` command above (or
include `~/.kube/config` in your `KUBECONFIG`).

It bind-mounts the repository into the k3d node at the same absolute path.
The example backend `hostPath` volume then exposes the local source to pods,
allowing builds in the running containers. Docker Desktop must have permission
to share this directory. If you move the checkout, delete/recreate the cluster
to update its bind mount:

```sh
k3d cluster delete shimshimney
./example/k3s/install.sh
```

Use `k3d cluster stop shimshimney` to stop the cluster without deleting it.
