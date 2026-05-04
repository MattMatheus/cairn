# Story: Cairnignore Workspace Filtering

## Metadata
- `id`: STORY-20260504-cairnignore-workspace-filtering
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-document-model-lifecycle, ADR-indexing-query-boundary]
- `success_metric`: Cairn consistently excludes `.cairnignore` paths from validation, indexing, search, and sync manifests.
- `release_scope`: required

## Problem Statement
- The north star says sync operates on the whole workspace except ignored paths, but Cairn does not yet define or apply `.cairnignore` rules across its workspace scanners.

## Scope
- In:
  - Parse `.cairnignore` with a small gitignore-style rule subset suitable for v1.
  - Apply ignore filtering to document discovery, workspace validation, local indexing, and sync manifest generation.
  - Keep `.cairn/` control metadata handled deliberately rather than accidentally hidden.
  - Add tests proving ignored invalid markdown does not block validation, indexing, or sync.
- Out:
  - Full Git-compatible ignore edge cases.
  - Per-command ignore overrides.
  - UI for editing ignore rules.

## Assumptions
- V1 can support comments, blank lines, directory suffixes, glob-like `*`, and negation only if cheap; otherwise negation can be deferred.

## Acceptance Criteria
1. `.cairnignore` rules are loaded from the workspace root.
2. Ignored markdown is skipped by validation and local index/search.
3. Ignored files are absent from sync manifests.
4. Tests cover ignored files, non-ignored files, and control metadata behavior.

## Validation
- Required checks:
  - Document/workspace validation tests.
  - Local index tests.
  - Sync manifest tests.
  - Full `go test ./...`.

## Dependencies
- `STORY-20260503-config-yaml-schema-validation`
- `STORY-20260504-sync-validation-gate`

## Risks
- Accidentally hiding `.cairn/sync-state.json` or generated control files from operations that still need them.

## Open Questions
- Resolved for V1: support `!negation` because existing validation/sync ignore parsing already handles it.

## Next Step
- Promote `STORY-20260504-sync-delete-purge-propagation`.

## Handoff Notes
- Engineering completed 2026-05-04.
- Confirmed validation and sync manifest generation already honored `.cairnignore`.
- Added `.cairnignore` filtering to local metadata indexing and full-text search scans.
- Added tests for ignored metadata and full-text documents.
- QA completed 2026-05-04 with focused workspace/localindex/syncstate tests and full `go test ./...`.
