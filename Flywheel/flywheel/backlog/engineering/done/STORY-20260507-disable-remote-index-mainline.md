# Story: Disable Remote Index Mainline

## Metadata
- `id`: STORY-20260507-disable-remote-index-mainline
- `owner_role`: Software Engineer
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: V1 search and index refresh succeed locally with no configured `remote_index.url`.
- `release_scope`: required

## Problem Statement
- Cairn v1 should not require a remote indexer service. Local `index_refresh` should rebuild SQLite, and `search_context` should return local metadata/full-text results while reporting semantic/remote modes as unavailable when requested.

## Scope
- In:
  - Verify and adjust runtime defaults so local index/search are the mainline behavior.
  - Ensure `index_status` and `index_refresh` language does not imply remote service dependency.
  - Ensure MCP and CLI behavior degrade clearly when semantic or remote modes are unavailable.
  - Add focused tests or update existing tests to lock the local-only path.
- Out:
  - Removing the remote index client entirely.
  - Implementing semantic search or CocoIndex.

## Assumptions
- Keeping the MCP tool names stable is preferred over removing index-related tools.
- Remote index code may remain behind optional config.
- `semantic` remains accepted by schema in v1 but should degrade clearly unless a rich-retrieval adapter is configured.

## Acceptance Criteria
1. With no `remote_index.url`, `cairn index refresh` rebuilds local SQLite and returns a successful local result.
2. With no `remote_index.url`, `cairn search --mode auto` uses local metadata/full-text only.
3. Explicit semantic/remote search requests return a clear unavailable/degraded response rather than failing opaquely.
4. MCP `index_status`, `index_refresh`, and `search_context` match the v1 semantics.
5. Tests cover local-only search/index behavior.

## Validation
- Required checks:
  - `go test ./...`
  - CLI smoke for `init`, `capture`, `index refresh`, and `search`.
- Additional checks:
  - MCP tool-call smoke for `search_context` and `index_refresh`.

## Dependencies
- `ARCH-20260507-core-v1-indexing-boundary-refresh` is accepted.
- Should follow or coordinate with `STORY-20260507-core-v1-docs-and-quickstart` so CLI/MCP wording matches docs.

## Risks
- Current code may already mostly behave this way; the work may be mostly tests and wording cleanup.

## Open Questions
- None for activation. PM selected graceful `semantic` degradation over hiding the mode from schema.

## PM Handoff
- `What changed`: Promoted into active as the second Cairn Core v1 story.
- `Why it matters`: It turns the architecture decision into tested runtime behavior.
- `Acceptance criteria`: Kept local-only CLI/MCP behavior as the measurable bar.
- `Risks and assumptions`: Do not remove optional remote index code unless it blocks local-first behavior.
- `Next state recommendation`: Engineering should implement after or alongside the docs story.

## Engineering Handoff
- `Implementation summary`: Updated local search `auto` mode so it attempts metadata and full-text only when no remote indexer is configured. Explicit `semantic` mode still degrades with warnings/unavailable modes. Updated `index_status` so local-only status is clean and points users to local search instead of warning about an unconfigured remote indexer. Updated `index_refresh` to rebuild local SQLite first, then optionally call a configured remote indexer.
- `Tests updated`: Adjusted local index, MCP ops, and CLI tests to lock local-only auto search, clean local-only index status, local-first index refresh, and no unconfigured remote warnings in CLI search.
- `Validation`: `go test ./...` passed. CLI smoke passed for `init`, `capture`, `index refresh`, and `search --mode auto` in `/private/tmp/cairn-local-only-smoke.nvZe44`. MCP smoke passed for `search_context` and `index_refresh`. Explicit CLI `search --mode semantic` returned a graceful warning without failing.
- `Action and approval notes`: Code/test edits are `local write`. `go test ./...` required approved escalation because the sandbox blocked Go build-cache writes under the user cache directory.
- `Open risks`: Existing optional remote index code remains in place; QA should focus on whether local-only mode is now quiet by default and explicit semantic mode still degrades clearly.
- `QA focus areas`: Verify acceptance criteria 1-5, especially that `auto` no longer reports semantic/remote degradation when no remote indexer is configured, while explicit `semantic` still does.
- `Next state recommendation`: Move to engineering QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed the local search auto-mode branch, `IndexStatus`, and `IndexRefresh` implementation. Confirmed `auto` only attempts metadata/full-text when no remote indexer is configured, explicit `semantic` still degrades, local-only `index_status` is clean, and `index_refresh` rebuilds local SQLite before optional remote work.
- `Evidence quality call`: Sufficient. QA reran focused tests with `go test ./internal/localindex ./internal/mcpops ./internal/cli`; engineering evidence included full `go test ./...`, CLI smoke, MCP smoke, and explicit semantic degradation smoke.
- `Defects`: None.
- `Required fixes`: None.
- `Residual risks`: Optional remote index code remains intentionally present; no-service sync smoke remains in the next story.
- `Next state recommendation`: Move to engineering done.

## Next Step
- Continue engineering with `STORY-20260507-local-blob-sync-smoke`.
