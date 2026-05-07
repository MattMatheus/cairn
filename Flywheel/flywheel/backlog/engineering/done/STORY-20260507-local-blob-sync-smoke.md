# Story: Local Blob Sync Smoke

## Metadata
- `id`: STORY-20260507-local-blob-sync-smoke
- `owner_role`: SDET
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: A single smoke path proves local docs, validation, local index/search, push, pull, and sync conflict visibility without CocoIndex services.
- `release_scope`: required

## Problem Statement
- The v1 confidence path should prove the Cairn Core loop without requiring Postgres, pgvector, CocoIndex, or a remote indexer. The existing local-dev smoke emphasizes the richer service stack.

## Scope
- In:
  - Add or update a smoke script for local filesystem remote-store emulation.
  - Cover init, capture, validate, index refresh, search, sync status, sync push, sync pull, and at least one conflict-status scenario if practical.
  - Keep the smoke script fast enough for routine developer use.
  - Keep Azurite/CocoIndex smoke as secondary or future/reference if it remains present.
- Out:
  - CocoIndex service smoke as the v1 confidence path.
  - Production Azure deployment validation.
  - Browser/UI testing.

## Assumptions
- Local filesystem remote-store emulation is acceptable as the fastest v1 confidence path.
- Azure Blob/Azurite coverage can remain as secondary integration validation.

## Acceptance Criteria
1. Smoke path runs without Docker or network services.
2. Smoke path creates at least two workspaces sharing one local remote store.
3. Smoke path proves a document can be pushed, pulled, indexed locally, and found through local search.
4. Smoke path reports or checks conflict behavior without mutating through divergence.
5. Documentation points new engineers at this smoke path first.

## Validation
- Required checks:
  - Run the new or updated smoke script.
  - `go test ./...`
- Additional checks:
  - Run Azurite sync smoke if local container runtime is available.

## Dependencies
- `STORY-20260507-disable-remote-index-mainline`

## Risks
- Conflict simulation can make smoke scripts brittle; keep it simple and deterministic.

## Open Questions
- Smoke location can be decided during implementation. PM preference is a core/no-service path that is easy to find from docs, whether under `deployments/local-dev/` or a clearer core-dev path.

## PM Handoff
- `What changed`: Promoted into active as the third Cairn Core v1 story.
- `Why it matters`: V1 needs one fast proof path that does not accidentally require the deferred rich-retrieval stack.
- `Acceptance criteria`: Kept no-Docker/no-network smoke as the bar.
- `Risks and assumptions`: Keep conflict checks deterministic; do not make CocoIndex smoke the default confidence path.
- `Next state recommendation`: Engineering/SDET should implement after local-only index/search behavior is locked.

## Engineering Handoff
- `Implementation summary`: Added `deployments/local-dev/core-smoke.sh`, a no-service Cairn Core smoke that uses `remote_sync.provider: local_fs`. Updated `README.md`, `docs/user/quickstart.md`, and `deployments/local-dev/README.md` to point new engineers at the no-service smoke first.
- `Smoke coverage`: The script initializes a primary workspace, writes local filesystem remote config, captures and validates a document, refreshes local SQLite, searches locally, pushes to the local remote store, creates a second workspace, pushes a remote-side document, pulls it into the second workspace, refreshes/searches there, and checks deterministic same-file divergence through `sync status`.
- `Validation`: `deployments/local-dev/core-smoke.sh` passed. `go test ./...` passed.
- `Action and approval notes`: Script/docs edits are `local write`. No Docker, network service, risky write, or production action was used.
- `Open risks`: The existing optional rich-retrieval smoke remains in `deployments/local-dev/smoke.sh`; docs label it optional.
- `QA focus areas`: Confirm the script does not start compose or require network services; confirm it creates multiple workspaces sharing one local remote store; confirm conflict status is checked without applying a divergent mutation; confirm docs point to this smoke first.
- `Next state recommendation`: Move to engineering QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reran `deployments/local-dev/core-smoke.sh`; it passed and printed a local temp workspace, local remote store path, and search phrase. Reran `go test ./...`; it passed. Reviewed the script behavior against all acceptance criteria.
- `Evidence quality call`: Sufficient. The smoke proves the no-service local filesystem remote-store path, local index/search after pull, and conflict-status reporting.
- `Defects`: None.
- `Required fixes`: None.
- `Residual risks`: Optional Azurite/CocoIndex smoke remains separate and was not run because it is outside the v1 no-service path.
- `Next state recommendation`: Move to engineering done.

## Next Step
- Continue engineering with `STORY-20260507-defer-cocoindex-artifacts`.
