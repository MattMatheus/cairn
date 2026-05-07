# Local Development

This directory now contains the Cairn Core v1 local development path. It uses visible markdown files, local SQLite search, and `remote_sync.provider: local_fs` for blob-style sync without Docker, cloud credentials, Postgres, pgvector, CocoIndex, Azurite, or a remote indexer.

## Core Smoke

Run the no-service confidence check from the repo root:

```sh
deployments/local-dev/core-smoke.sh
```

The script creates two throwaway workspaces sharing one local filesystem remote store. It checks:

- `init`
- `capture`
- `validate`
- local `index refresh`
- local `search --mode auto`
- `sync push`
- `sync pull`
- conflict detection without applying a divergent mutation

## Manual Loop

Create a workspace:

```sh
WORK_ROOT="$(mktemp -d)/cairn-core"
go run ./cmd/cairn --root "$WORK_ROOT" init --workspace-id cairn:workspace:core-dev
```

Add a local filesystem remote store to `$WORK_ROOT/.cairn/config.yaml`:

```sh
go run ./cmd/cairn --root "$WORK_ROOT" setup local-sync --remote-root /tmp/cairn-local-remote
```

Exercise the local-first loop:

```sh
go run ./cmd/cairn --root "$WORK_ROOT" capture \
  --actor codex \
  --title "Core Dev Note" \
  --type note \
  --body "The Cairn Core path uses local search and blob-style sync."

go run ./cmd/cairn --root "$WORK_ROOT" validate
go run ./cmd/cairn --root "$WORK_ROOT" index refresh
go run ./cmd/cairn --root "$WORK_ROOT" search --query "local search"
go run ./cmd/cairn --root "$WORK_ROOT" sync dry-run
go run ./cmd/cairn --root "$WORK_ROOT" sync push
```

Sync refuses divergent local and remote changes instead of merging silently.

## Config Sample

`workspace-config.local_fs.yaml` is the sample for this path. It intentionally omits `remote_index.url` so `search --mode auto` stays local-only.

Generated smoke workspaces and remote stores are created under the system temp directory unless `WORK_ROOT` is set.

## Pilot Readiness

Run the broader preflight before inviting an engineer into a first pilot:

```sh
deployments/local-dev/pilot-check.sh
```

This runs the Go test suite, builds a throwaway binary, validates/searches `examples/pilot-workspace`, and then runs `core-smoke.sh`.
