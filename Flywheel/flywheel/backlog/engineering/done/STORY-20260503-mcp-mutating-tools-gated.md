# Story: MCP Mutating Tools Gated Surface

## Metadata
- `id`: STORY-20260503-mcp-mutating-tools-gated
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface, ADR-document-model-lifecycle, ADR-sync-conflict-behavior]
- `success_metric`: MCP can expose local write tools behind an explicit opt-in server mode while preserving delete/purge exclusions.
- `release_scope`: required

## Problem Statement
- Local mutation operation adapters exist, but the MCP transport currently exposes only read-only tools for safety.

## Scope
- In:
  - Add explicit opt-in server mode for local write MCP tools.
  - Register `capture_note`, `promote_document`, and `archive_document`.
  - Preserve common envelope responses and next steps.
  - Ensure hard delete/purge remains absent.
  - Add tests for read-only versus write-enabled registration.
- Out:
  - `sync_pull`/`sync_push` mutation exposure until sync apply stories land.
  - Hard delete/purge.
  - Remote profile write behavior.

## Assumptions
- Write-enabled MCP server should be opt-in, not the default.

## Acceptance Criteria
1. Default MCP server remains read-only.
2. Write-enabled mode registers local lifecycle mutation tools.
3. Mutation handlers call `internal/mcpops` adapters.
4. Purge/delete is absent in all modes.
5. Tests cover registration and representative mutation call.

## Validation
- Required checks:
  - Unit/integration tests for MCP registration and handlers.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against MCP ADR safety constraints.

## Dependencies
- `STORY-20260503-mcp-readonly-server-transport`
- `STORY-20260503-mcp-lifecycle-mutation-adapters`

## Risks
- Keep opt-in explicit enough that agents do not get writes accidentally.

## Open Questions
- CLI flag name for write-enabled MCP mode.

## Engineering Handoff
- Implemented 2026-05-03.
- Added `mcpserver.WithLocalWrites()` as explicit opt-in server configuration.
- Default `mcpserver.New(local)` remains read-only.
- Write-enabled mode registers `capture_note`, `promote_document`, and `archive_document`.
- Mutation handlers call the existing `internal/mcpops` lifecycle adapters.
- Added CLI mode `cairn mcp local-writes` while preserving `cairn mcp readonly`.

## QA Handoff
- Accepted 2026-05-03.
- Verified default MCP server registers read-only tools only.
- Verified write-enabled mode registers lifecycle mutation tools.
- Verified representative `capture_note` call returns the common mutation envelope and writes the expected document.
- Verified sync/index mutations remain absent from this lifecycle-gated mode.
- Verified delete/purge-shaped tools are absent in all modes.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260503-cocoindex-local-packaging`.
