# Story: Sync Pull Apply

## Metadata
- `id`: STORY-20260503-sync-pull-apply
- `owner_role`: Software Architect
- `status`: done
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

## Engineering Handoff
- Implemented 2026-05-03.
- Added `syncstate.ApplyPull` for safe remote-only pull plans.
- Pull apply fetches create/edit/move/archive content through a remote object reader.
- Pull apply removes local files for remote deletes only after the plan is accepted as remote-only.
- Sync state advances to the accepted remote manifest only after all pull writes/removals succeed.
- Added `mcpops.Local.SyncPull` and a CLI `sync pull` surface.

## QA Handoff
- Accepted 2026-05-03.
- Verified remote create/edit application.
- Verified remote archive/move preserves document identity by path and `document_id`.
- Verified remote delete removes local files only on accepted remote-only plans.
- Verified divergence refuses without changing workspace files or sync state.
- Verified missing remote objects do not advance sync state.
- Fixed QA finding: move/archive now fetches and writes the new remote object before removing the old local path.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260503-sync-push-apply`.
