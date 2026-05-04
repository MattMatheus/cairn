# Cairn Quickstart

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

Search uses local metadata and full-text lookup. If a remote indexer is configured, semantic/remote search can participate; otherwise Cairn reports degraded remote modes and remains useful locally.

## Sync

Use dry-run before mutating remote or local state:

```sh
cairn sync status
cairn sync dry-run
cairn sync push
cairn sync pull
```

Sync refuses divergent local/remote changes instead of merging. Resolve conflicts manually, then rerun status and dry-run.

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
