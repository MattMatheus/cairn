# Story: Sync Pull Apply

## Metadata
- `id`: STORY-20260503-sync-pull-apply
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-mcp-operation-surface]
- `success_metric`: Cairn can safely apply remote-only workspace changes to the local workspace and update local sync state.
- `release_scope`: required

## Problem Statement
- Cairn can detect remote-only changes but cannot yet apply them locally.

## Scope
- In:
  - Implement pull apply for safe remote-only plans.
  - Fetch remote objects through the remote store adapter.
  - Create/edit/move/archive/delete local files according to the accepted remote manifest.
  - Update `.cairn/sync-state.json` only after successful apply.
  - Add CLI/MCP operation wiring where practical.
- Out:
  - Push apply.
  - Merge conflict resolution.
  - Live Azure integration tests.

## Assumptions
- Pull must refuse if status is diverged or if local changes exist outside a safe plan.

## Acceptance Criteria
1. Remote-only creates/edits are applied locally.
2. Remote moves/archives preserve document identity and paths.
3. Remote deletes remove local files only when safe.
4. Divergence refuses without changing files or state.
5. Sync state updates only after a successful pull.

## Validation
- Required checks:
  - Unit tests with fake remote store.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against sync ADR mutation refusal rules.

## Dependencies
- `STORY-20260503-sync-dry-run-plan`

## Risks
- Partial failure handling needs care; state must not advance before all file writes succeed.

## Open Questions
- Whether pull should stage writes through temporary files for stronger atomicity in v1.

## Next Step
- Engineering should implement only after dry-run planning lands.

