# Story: Sync Dry-Run Plan

## Metadata
- `id`: STORY-20260503-sync-dry-run-plan
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-mcp-operation-surface]
- `success_metric`: Cairn can produce a non-mutating pull/push plan from sync status and remote store state.
- `release_scope`: required

## Problem Statement
- Sync status reports divergence, but users need to see exactly what pull or push would do before any mutating sync operation is enabled.

## Scope
- In:
  - Add reusable dry-run sync planning for pull and push.
  - Use local manifest, local sync state, and remote manifest/store adapter.
  - Report planned creates, edits, moves, archives, deletes, and refusal conflicts.
  - Surface next steps for safe pull/push or manual reconciliation.
  - Add CLI/MCP-facing data shape where practical without enabling mutation.
- Out:
  - Applying local file changes.
  - Writing remote objects.
  - Updating sync state.
  - Automatic merge or conflict resolution.

## Assumptions
- Existing sync status comparison remains the safety gate.

## Acceptance Criteria
1. Clean workspaces produce an empty plan.
2. Local-only changes produce a push plan.
3. Remote-only changes produce a pull plan.
4. Diverged workspaces produce a refusal plan with conflicts.
5. Dry-run never mutates files, state, or remote objects.

## Validation
- Required checks:
  - Unit tests for clean, local-only, remote-only, and diverged plans.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against sync ADR refusal requirements.

## Dependencies
- `STORY-20260503-sync-status-conflict-report`
- `STORY-20260503-azure-blob-sync-adapter`

## Risks
- Keep plan output stable enough for CLI/MCP without over-designing full sync execution.

## Open Questions
- Exact plan response shape for future MCP `sync_pull`/`sync_push` dry-run behavior.

## Next Step
- Engineering should implement dry-run planning before mutating sync operations.

