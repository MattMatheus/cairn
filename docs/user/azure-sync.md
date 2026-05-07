# Azure Blob Sync

Azure Blob is the v1 multiplayer backend path. The pilot shape is intentionally simple: one Blob container per pod, provisioned by infra, with engineers authenticating through Azure CLI.

## Assumptions

- Infra creates the storage account and one container for the pod.
- Engineers have Azure CLI installed and can run `az login`.
- Cairn uses the current Azure CLI identity for Blob access.
- The container holds Cairn workspace documents and sync metadata.
- Soft delete, blob versioning, and lifecycle policy are not required for the first pilot.
- The current remote manifest conflict refusal is acceptable for small pod size.

## Configure A Workspace

```sh
WORK_ROOT="$HOME/CairnPilot"
mkdir -p "$WORK_ROOT"

cairn --root "$WORK_ROOT" setup azure-sync \
  --account cairnpodstorage \
  --container pod-a

az login
cairn --root "$WORK_ROOT" doctor --remote
cairn --root "$WORK_ROOT" sync status
```

If infra gives you a Blob endpoint instead of an account name:

```sh
cairn --root "$WORK_ROOT" setup azure-sync \
  --endpoint https://cairnpodstorage.blob.core.windows.net \
  --container pod-a
```

## Container And Prefix

For v1, use one container per pod and leave `prefix` empty. That keeps the remote layout easy to inspect and makes the container itself the pod boundary.

`workspace_id` remains useful even with an empty prefix. It identifies the local workspace in manifests and provenance output. If Cairn later supports multiple workspaces inside a shared container, `prefix` can become the pod or workspace slug without changing the sync model.

## Daily Loop

```sh
cairn --root "$WORK_ROOT" doctor --remote
cairn --root "$WORK_ROOT" sync status
cairn --root "$WORK_ROOT" sync dry-run
cairn --root "$WORK_ROOT" sync pull
cairn --root "$WORK_ROOT" sync push
```

Use `sync dry-run` before mutating local or remote state. If local and remote have both changed since the last accepted base, Cairn refuses the operation instead of merging.
