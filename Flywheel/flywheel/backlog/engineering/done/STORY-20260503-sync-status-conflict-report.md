# Story: Sync Status And Conflict Refusal Report

## Metadata
- `id`: STORY-20260503-sync-status-conflict-report
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-mcp-operation-surface]
- `success_metric`: Cairn can report local/remote sync status and conflict refusal details without mutating local or remote state.
- `release_scope`: required

## Problem Statement
- Sync manifest comparison exists, but users and agents need a read-only status operation before any pull/push implementation.

## Scope
- In:
  - Implement reusable `sync_status` operation from local manifest, local state, and supplied/loaded remote manifest fixture.
  - Return local changes, remote changes, divergence/conflict details, warnings, and next steps.
  - Ensure refused/diverged status does not update state or files.
  - Add tests for clean, local-only, remote-only, and diverged states.
- Out:
  - Azure Blob API calls.
  - `sync_pull` or `sync_push` mutation.
  - Automatic merge/conflict resolution.

## Assumptions
- Remote manifest can be represented as local JSON until Azure adapter lands.

## Acceptance Criteria
1. Clean status reports no changes.
2. Local-only and remote-only changes are classified.
3. Divergence reports conflicts and refusal next steps.
4. Status never mutates sync state or workspace files.
5. Tests cover safe and refused cases.

## Validation
- Required checks:
  - Unit/integration tests for sync status behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against sync ADR.

## Dependencies
- `STORY-20260502-sync-manifest-state`

## Risks
- Keep remote loading swappable for Azure Blob later.

## Open Questions
- Resolved for local profile: load the remote manifest fixture from `.cairn/remote-manifest.json`, matching the sync ADR path.

## Next Step
- Engineering should implement the read-only sync status operation and tests.

## PM Handoff
- Promoted on 2026-05-03 after the local CLI surface landed.
- Keep this story read-only: no pull, push, state updates, or file writes.
- Make the remote manifest loader easy to replace when the Azure Blob adapter lands.

## Engineering Handoff
- Implemented reusable `syncstate.StatusReport` from local manifest, local sync state, and `.cairn/remote-manifest.json`.
- Added sync status envelope mapping with local changes, remote changes, conflicts, divergence, warnings, unavailable modes, and next steps.
- Added `Local.SyncStatus` MCP operation adapter.
- Added `cairn sync status` to the local CLI.
- Filtered local sync control files from status comparison so the report does not treat `.cairn/sync-state.json` or `.cairn/remote-manifest.json` as user workspace changes.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/syncstate ./internal/mcpops ./internal/cli`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" sync status`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed clean, local-only, remote-only, missing-remote, and diverged status cases.
- Confirmed diverged status returns conflicts, refusal warnings, and manual recovery next steps.
- Confirmed status does not update local sync state or overwrite workspace files.
- Confirmed MCP and CLI surfaces expose the read-only status result.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote the Azure Blob remote sync adapter story so the local remote-manifest fixture can be replaced with a real remote manifest source.
