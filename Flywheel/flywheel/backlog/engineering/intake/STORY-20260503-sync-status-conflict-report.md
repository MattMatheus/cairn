# Story: Sync Status And Conflict Refusal Report

## Metadata
- `id`: STORY-20260503-sync-status-conflict-report
- `owner_role`: Software Architect
- `status`: intake
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
- Exact local path for fixture/remote manifest loading in local profile.

## Next Step
- PM should promote before sync pull/push work.
