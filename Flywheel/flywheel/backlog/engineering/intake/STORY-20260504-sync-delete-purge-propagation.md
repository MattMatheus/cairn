# Story: Sync Delete And Purge Propagation

## Metadata
- `id`: STORY-20260504-sync-delete-purge-propagation
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-document-model-lifecycle]
- `success_metric`: Sync treats local deletions from explicit purge as manifest changes without exposing remote hard delete through MCP lifecycle tools.
- `release_scope`: required

## Problem Statement
- Purge now removes archived local files, but sync push/pull behavior around deleted paths needs explicit tests and operator-facing output.

## Scope
- In:
  - Verify local deletion appears as a planned sync change from the last accepted base.
  - Apply push-side deletion to the remote store only through sync push.
  - Apply pull-side remote deletion locally only when safe.
  - Preserve conflict refusal when local and remote both changed the same path since base.
  - Report next steps after deletion plans/applications.
- Out:
  - MCP purge/delete tools.
  - Retention/tombstone policy beyond manifest state.
  - Per-document sync.

## Assumptions
- V1 can use absence from the next accepted manifest as the delete signal.

## Acceptance Criteria
1. `sync dry-run` reports local deletions after a CLI purge.
2. `sync push` removes remote blobs for safe deletion changes.
3. `sync pull` removes local files for safe remote deletion changes.
4. Conflict detection refuses ambiguous local/remote edits without updating state.
5. MCP still has no purge/delete lifecycle tool.

## Validation
- Required checks:
  - Syncstate unit tests.
  - CLI sync dry-run/push/pull tests if command output changes.
  - MCP registration tests if tool surfaces are touched.
  - Full `go test ./...`.

## Dependencies
- `STORY-20260504-cli-purge-archived-document`
- `STORY-20260504-sync-validation-gate`

## Risks
- Remote deletion must remain a sync operation, not an agent lifecycle primitive.

## Open Questions
- Whether remote stores should hard-delete blobs immediately or retain provider-level versions when available.

## Next Step
- Engineering should inspect current sync delete handling before deciding whether this is a bug fix or a test-hardening story.
