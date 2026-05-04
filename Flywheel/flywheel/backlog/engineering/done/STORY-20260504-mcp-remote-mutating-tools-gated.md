# Story: MCP Remote Mutating Tools Gated Surface

## Metadata
- `id`: STORY-20260504-mcp-remote-mutating-tools-gated
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-mcp-operation-surface, ADR-sync-conflict-behavior, ADR-indexing-query-boundary]
- `success_metric`: MCP can expose safe sync and index mutation tools only behind an explicit opt-in mode.
- `release_scope`: required

## Problem Statement
- `sync_pull`, `sync_push`, and `index_refresh` operation adapters exist, but MCP currently exposes only read-only tools and local lifecycle writes.

## Scope
- In:
  - Add an explicit remote-write MCP server option separate from read-only and local lifecycle writes.
  - Register `sync_pull`, `sync_push`, and `index_refresh` only in that mode.
  - Preserve default read-only behavior.
  - Preserve `local-writes` lifecycle-only behavior.
  - Add CLI MCP mode for the remote-write surface if the name is sufficiently explicit.
  - Add tests for registration boundaries and representative handler responses.
- Out:
  - Hard delete/purge exposure.
  - Automatic sync scheduling.
  - Live Azure integration tests.

## Assumptions
- The remote-write MCP mode should be harder to invoke accidentally than local lifecycle writes.

## Acceptance Criteria
1. Default MCP server remains read-only.
2. Local write MCP mode remains limited to lifecycle mutations.
3. Remote-write mode registers `sync_pull`, `sync_push`, and `index_refresh`.
4. Delete/purge remains absent in all MCP modes.
5. Tests cover registration boundaries and at least one remote mutation handler.

## Validation
- Required checks:
  - MCP server unit tests.
  - Repository formatting/lint checks if configured.

## Dependencies
- `STORY-20260503-sync-pull-apply`
- `STORY-20260503-sync-push-apply`
- `STORY-20260503-index-refresh-contract-wiring`
- `STORY-20260503-mcp-mutating-tools-gated`

## Risks
- Avoid making remote mutation tools available by default.

## Open Questions
- Exact CLI mode name: candidate `cairn mcp remote-writes`.

## Engineering Handoff
- Implemented 2026-05-04.
- Added `mcpserver.WithRemoteWrites()` as explicit opt-in server configuration.
- Added `cairn mcp remote-writes`.
- Remote-write mode registers `sync_pull`, `sync_push`, and `index_refresh`.
- Default read-only and local lifecycle write modes remain separate.

## QA Handoff
- Accepted 2026-05-04.
- Verified default MCP server remains read-only.
- Verified local write mode remains limited to lifecycle mutations.
- Verified remote-write mode registers sync/index mutation tools.
- Verified delete/purge remain absent in all modes.
- Verified representative `index_refresh` handler response.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260504-remote-profile-config-client-wiring`.
