# ADR: Sync And Conflict Behavior

## Status

Accepted

## Context

Cairn v1 uses on-demand workspace sync through CLI and MCP. Azure Blob storage is the first shared backend, with one storage account per pod for isolation, cost attribution, and simpler operations. Cloud drive sync, OneDrive, Git-backed sync, automatic merge tooling, and cross-pod federation are out of scope.

The sync system must preserve the file-first model while avoiding silent corruption when local and remote work diverge.

## Decision

Sync operates on the whole Cairn workspace except ignored paths. Ignore rules live in `.cairnignore` and use gitignore-style syntax.

The remote Blob layout mirrors the local workspace root, optionally under a configured prefix:

```yaml
remote:
  provider: azure_blob
  container: cairn
  prefix: ""
```

Local sync state lives at:

```text
/.cairn/sync-state.json
```

Remote manifest lives at:

```text
/.cairn/remote-manifest.json
```

The remote manifest records, at minimum:

| Field | Purpose |
| --- | --- |
| `manifest_version` | Schema version for the manifest |
| `generated_at` | Manifest generation timestamp |
| `workspace_id` | Pod/workspace identity when configured |
| `entries[].path` | Workspace-relative path |
| `entries[].kind` | `file` or future supported object kind |
| `entries[].size` | Byte size |
| `entries[].hash` | Content hash |
| `entries[].modified` | Last modified time observed by Cairn |
| `entries[].document_id` | Cairn document id when available |
| `entries[].status` | Document status when available |
| `entries[].type` | Document type when available |

Local sync state records, at minimum:

| Field | Purpose |
| --- | --- |
| `state_version` | Schema version for local state |
| `last_remote_manifest_hash` | Hash of the last accepted remote manifest |
| `last_sync_at` | Last accepted sync timestamp |
| `entries` | Last accepted remote entries used as the divergence base |

The last accepted remote manifest is the base for divergence detection.

Sync treats file content changes, creates, moves, archives, and deletes as manifest changes. Moves and archives should preserve document id when frontmatter is available. A path change with the same document id is a move. An archived document is a move under `/archive` plus `status: archived`.

If local and remote both changed since the last accepted base, Cairn refuses the operation rather than merging. A refused sync must not:

- update local sync state
- overwrite local files
- overwrite remote files
- publish a new remote manifest

User-facing refusal output should include the conflicting paths, detected local and remote changes, and suggested recovery actions. Recovery is manual: pull, archive, rename, edit, or otherwise reconcile files and retry.

`sync_status` is read-only. `sync_pull` and `sync_push` are mutating operations. MCP may expose `sync_status`, `sync_pull`, and `sync_push`, but purge/delete remains CLI-only.

Document sync requires valid core frontmatter for Cairn-managed markdown unless the file is explicitly ignored. Index artifacts are not normal document sync by default; index refresh publishes or updates derived artifacts separately.

Git compatibility is limited to light nudging. Git is not the Cairn backend.

## Alternatives Considered

- Git-backed sync. Rejected because Git is not the product backend and would force teams into a workflow Cairn is meant to avoid.
- Cloud-drive sync. Rejected for v1 because it is explicitly out of scope.
- Automatic three-way merge. Rejected for v1 because markdown and frontmatter merge errors would risk corrupting canonical knowledge.
- Per-document sync. Deferred because whole-workspace sync is simpler for small pods.
- Path-only manifests. Rejected because document ids are needed to understand moves and archives.

## Consequences

Conflict refusal makes sync safer and easier to reason about, but users must resolve divergence manually.

Manifest and local state schemas become core compatibility surfaces and should be versioned from the start.

Renames, archive moves, and ADR number allocation rely on durable document ids.

## Follow-On Implementation Paths

- Implement `.cairnignore` parsing.
- Implement remote manifest generation and validation.
- Implement local sync state read/write.
- Implement `sync_status`, `sync_pull`, and `sync_push`.
- Implement conflict refusal reporting.
- Add tests for creates, edits, moves, archives, deletes, and refused divergence.
