# Story: Stop sharing local workspace config via sync

## Metadata
- `id`: STORY-20260508-workspace-config-not-shared
- `owner_role`: Software Architect
- `status`: ready
- `source`: planning
- `decision_refs`: [ARCH-20260502-sync-conflict-behavior]
- `success_metric`: `.cairn/config.yaml` and `.cairn/repos.yaml` are not transmitted by `syncstate.push`; pull tolerates remote stores that still carry stale entries.
- `release_scope`: required

## Problem Statement
- `internal/syncstate/manifest.go:60-95` and `internal/workspace/init.go:197-203` together cause sync push to upload `.cairn/config.yaml` and `.cairn/repos.yaml`. These are user/host-specific (Azure auth, local repo paths) and should never propagate to collaborators (High, H3).

## Scope
- In:
  - Extend `statusComparableEntries` in `internal/syncstate/status.go:91-102` to filter `.cairn/config.yaml` and `.cairn/repos.yaml`.
  - Update default `.cairnignore` in `internal/workspace/init.go:197-203` to document the policy.
  - Pull-side: ignore (do not delete locally, do not overwrite) any remote-manifest entries for the filtered paths.
  - Migration note in `docs/user/` for existing pilots.
- Out:
  - Wholesale rework of what lives in `.cairn/`.

## Assumptions
- No collaborator workflow currently relies on shared `config.yaml`/`repos.yaml`.
- `.cairn/schemas/` remains shared (workspace-durable).

## Acceptance Criteria
1. Push test: workspace with `config.yaml` and `repos.yaml` populated produces a manifest that does NOT include those paths.
2. Pull test: a remote manifest carrying those paths does not overwrite the local copies.
3. Sync-state status output does not flag these files as "ahead" or "behind".
4. Existing tests around `.cairn/index/` and `.cairn/sync-state.json` filtering still pass.

## Validation
- Required checks:
  - `go test ./internal/syncstate/...`
  - `go test ./internal/workspace/...`
  - `make pilot-check`
- Additional checks:
  - Manual: two-workspace local-fs sync demo to confirm config not shared.

## Dependencies
- None.

## Risks
- Pilots with stale remote entries see no immediate cleanup; document an opt-in cleanup command.

## Open Questions
- Add `cairn sync prune --shared-config` cleanup command? (Out of scope unless trivial.)

## Next Step
- PM ranks; engineering implements.
