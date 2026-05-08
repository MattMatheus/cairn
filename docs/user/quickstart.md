# Cairn Quickstart

This quickstart covers Cairn Core v1: local markdown files, validation, local SQLite search, MCP access, and optional blob-backed sync. It does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer.

## Install

Install a released binary into userland:

```sh
curl -fsSL https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
cairn version
```

For source contributors, `go run ./cmd/cairn version` works from the repository root. Pilots should use the released binary path above.

## Create A Workspace

```sh
WORK_ROOT="$HOME/CairnPilot"
mkdir -p "$WORK_ROOT"
cd "$WORK_ROOT"
cairn init
cairn doctor
cairn doctor --full
cairn validate
```

`init` creates the standard folders, `.cairn/config.yaml`, `.cairnignore`, starter schemas, onboarding docs, and terse `AGENTS.md` / `CLAUDE.md` pointers. Existing files are preserved.
If you accidentally run `cairn init` in the Cairn source repository, Cairn refuses unless `--force` is provided.
Use `doctor --full` for a fuller pilot readiness summary across validation, index/search, sync, remote configuration, and MCP tool surfaces.

For a local no-service sync pilot, use the setup helper instead of editing config by hand:

```sh
cairn setup local-sync --remote-root /tmp/cairn-local-remote
```

This runs workspace initialization if needed and writes `remote_sync.provider: local_fs` to `.cairn/config.yaml`.

## Capture A Note

For a short human-friendly capture path:

```sh
cairn note --title "Auth Timeout Investigation" --type investigation --body "Initial notes about the timeout."
```

`note` uses `CAIRN_ACTOR`, `USER`, or `USERNAME` as the actor when `--actor` is not supplied.

For prompt-driven capture:

```sh
cairn capture --interactive
```

End the interactive body with a line containing only `.`.

For explicit agent or automation capture:

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

Azure Blob is the shared pod backend path. See [Azure Blob Sync](azure-sync.md) for the one-container-per-pod setup.

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

## Pilot Helpers

See [Pilot Helpers](pilot-helpers.md) for:

- one-command readiness checks with `cairn doctor --full`
- repo attachment and `.cairn-workspace` discovery
- ADO PR completion candidate capture
- the VS Code command-palette helper
- local markdown health reports with `cairn health report`
