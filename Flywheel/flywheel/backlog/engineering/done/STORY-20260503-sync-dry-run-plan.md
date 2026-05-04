# Story: Sync Dry-Run Plan

## Metadata
- `id`: STORY-20260503-sync-dry-run-plan
- `owner_role`: Software Architect
- `status`: done
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
- Resolved for v1: plan shape includes direction, safe/refused flag, planned changes, conflicts, warnings, and next steps so pull/push can reuse it.

## Next Step
- Engineering should implement dry-run planning before mutating sync operations.

## PM Handoff
- Promoted on 2026-05-03 as the first story in the new sync batch.
- Keep the operation strictly non-mutating.
- Prefer reusable plan data over CLI-only prose so future pull/push stories can consume it.

## Engineering Handoff
- Added reusable `syncstate.Plan` and `PlanFromStatus`.
- Added `SyncPlanData` schema shape with direction, safe flag, planned changes, conflicts, and diverged state.
- Added `syncstate.PlanEnvelope` for CLI/MCP-style dry-run responses.
- Added `Local.SyncDryRun` operation adapter.
- Added `cairn sync dry-run`.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/syncstate ./internal/mcpops ./internal/cli`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" sync dry-run`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed clean, push, pull, and refused plan directions.
- Confirmed diverged workspaces return conflicts and refusal warnings.
- Confirmed dry-run does not mutate sync state or workspace files.
- Confirmed CLI and operation adapter expose reusable plan data.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote `STORY-20260503-sync-pull-apply` so remote-only plans can be safely applied locally.
