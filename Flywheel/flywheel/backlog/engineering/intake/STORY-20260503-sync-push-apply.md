# Story: Sync Push Apply

## Metadata
- `id`: STORY-20260503-sync-push-apply
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-mcp-operation-surface]
- `success_metric`: Cairn can safely publish local-only workspace changes to the remote store and update manifests/state.
- `release_scope`: required

## Problem Statement
- Cairn can detect local-only changes but cannot yet publish them to the remote store.

## Scope
- In:
  - Implement push apply for safe local-only plans.
  - Write changed workspace objects through the remote store adapter.
  - Publish remote manifest only after object writes succeed.
  - Update local sync state only after successful remote manifest publication.
  - Add CLI/MCP operation wiring where practical.
- Out:
  - Pull apply.
  - Automatic merge.
  - Live Azure integration tests.

## Assumptions
- Push must refuse if remote changed since the accepted base.

## Acceptance Criteria
1. Local creates/edits are written to remote.
2. Local moves/archives/deletes update the remote object set and manifest.
3. Divergence refuses without writing remote objects, remote manifest, or local state.
4. Remote manifest publication is last remote write.
5. Sync state updates only after successful push.

## Validation
- Required checks:
  - Unit tests with fake remote store.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against sync ADR mutation refusal rules.

## Dependencies
- `STORY-20260503-sync-dry-run-plan`

## Risks
- Remote delete semantics should remain explicit and well-tested.

## Open Questions
- Whether v1 should keep deleted remote objects as tombstones or hard-delete blobs.

## Next Step
- Engineering should implement after dry-run planning and preferably after pull apply.

