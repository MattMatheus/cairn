# Cairn Quickstart

This quickstart covers Cairn Core v1: local markdown files, validation, local SQLite search, MCP access, and optional blob-backed sync. It does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer.

## Install

From the repository root:

```sh
go build -o ./bin/cairn ./cmd/cairn
```

Use `./bin/cairn` in the commands below, or put `bin/` on your `PATH`.

## Create A Workspace

```sh
cairn init
cairn validate
```

`init` creates the standard folders, `.cairn/config.yaml`, `.cairnignore`, starter schemas, onboarding docs, and terse `AGENTS.md` / `CLAUDE.md` pointers. Existing files are preserved.

For a local no-service sync pilot, use the setup helper instead of editing config by hand:

```sh
cairn setup local-sync --remote-root /tmp/cairn-local-remote
```

This runs workspace initialization if needed and writes `remote_sync.provider: local_fs` to `.cairn/config.yaml`.

## Capture A Note

```sh
cairn capture \
  --actor codex \
  --title "Auth Timeout Investigation" \
  --type investigation \
  --body "Initial notes about the timeout."
```

For longer content:

```sh
cairn capture --actor codex --title "Deploy Handoff" --type handoff --body-file notes.md
```

Or stdin:

```sh
cairn capture --actor codex --title "Deploy Handoff" --type handoff --body-file -
```

## Promote Knowledge

Promotion to `proposed` is the review stage:

```sh
cairn promote agents/codex/auth-timeout-investigation.md --type investigation
```

Promotion to `canonical` requires valid managed frontmatter and must come from `proposed`:

```sh
cairn promote working/auth-timeout-investigation.md --status canonical
```

Decision documents promoted to canonical receive ADR-style numbering under `decisions/`.

## Archive Or Purge

Archive keeps history:

```sh
cairn archive decisions/ADR-0001-old-choice.md
```

Purge is hard delete and CLI-only. It only works for archived documents and requires explicit confirmation:

```sh
cairn purge --confirm-purge archive/decisions/ADR-0001-old-choice.md
```

## Search And Index

```sh
cairn search "auth timeout"
cairn index status
cairn index refresh
```

Search uses local metadata and full-text lookup. `index refresh` rebuilds the local SQLite index. Semantic or remote search is deferred rich-retrieval work; if such a mode is requested without an optional adapter configured, Cairn should report a clear degraded/unavailable mode and keep local search usable.

## Sync

Remote sharing uses a blob-style store plus a remote manifest. For the no-service local path, run `cairn setup local-sync --remote-root /tmp/cairn-local-remote`.

Use dry-run before mutating remote or local state:

```sh
cairn sync status
cairn sync dry-run
cairn sync push
cairn sync pull
```

Sync refuses divergent local/remote changes instead of merging. A refused sync should not overwrite local files, overwrite remote objects, update local sync state, or publish a new remote manifest. Resolve conflicts manually, then rerun status and dry-run.

Azure Blob can also be used through `remote_sync.provider: azure_blob`, but it is an advanced integration path for v1, not required for the default quickstart.

For a repeatable no-service sync check from a repo checkout, run:

```sh
deployments/local-dev/core-smoke.sh
```

## MCP Modes

```sh
cairn mcp readonly
cairn mcp local-writes
cairn mcp remote-writes
```

- `readonly`: read/search/validate/bootstrap surfaces.
- `local-writes`: capture, promote, and archive.
- `remote-writes`: sync pull, sync push, and index refresh.

Hard delete/purge is never exposed through MCP.
