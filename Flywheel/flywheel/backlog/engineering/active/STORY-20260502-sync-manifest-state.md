# Story: Sync Manifest And Local State Foundation

## Metadata
- `id`: STORY-20260502-sync-manifest-state
- `owner_role`: Software Architect
- `status`: active
- `source`: direct
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-document-model-lifecycle]
- `success_metric`: Cairn can build and compare local sync manifests/state enough to detect creates, edits, moves, archives, deletes, and divergence before any remote write occurs.
- `release_scope`: required

## Problem Statement
- Sync safety depends on deterministic manifest and local state behavior. Cairn needs this foundation before implementing Azure Blob push/pull.

## Scope
- In:
  - Parse `.cairnignore` using gitignore-style behavior.
  - Generate workspace manifest entries for non-ignored files.
  - Include path, kind, size, hash, modified time, document id, status, and type when available.
  - Read and write `/.cairn/sync-state.json`.
  - Compare current local manifest, last accepted base, and a supplied remote manifest.
  - Detect creates, edits, moves, archives, deletes, and local/remote divergence.
  - Add tests for manifest generation and divergence detection.
- Out:
  - Azure Blob API calls.
  - Actual `sync_pull` or `sync_push` remote mutation.
  - Automatic merge or conflict resolution.

## Assumptions
- Document validation and metadata parsing are available from the document model story.
- Capture, promotion, and archive operations are available from `STORY-20260502-capture-promotion-archive`.
- Remote manifest loading can be represented by local fixtures in this story.
- This story should create the smallest sync-focused package boundary needed for manifest/state comparison while reusing `internal/document` parsing and validation.

## Acceptance Criteria
1. Cairn can generate a manifest for a fixture workspace while honoring `.cairnignore`.
2. Manifest entries include document metadata when frontmatter is available.
3. Local sync state can store the last accepted remote manifest hash and entries.
4. Change detection classifies create, edit, move, archive, and delete cases.
5. Divergence detection refuses cases where local and remote both changed since base.
6. Tests cover safe and refused sync comparison cases.

## Validation
- Required checks:
  - Unit tests for ignore handling, manifest generation, state persistence, and divergence detection.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-sync-conflict-behavior.md`.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`

## Risks
- Modified timestamps can be unstable across filesystems; hashes must be the primary content signal.
- Move detection without document ids may be ambiguous and should degrade conservatively.

## Open Questions
- Resolved for this slice: store the full normalized base manifest in local sync state for debuggability and deterministic comparison. A compact representation can be introduced later if size becomes a problem.

## Next Step
- Engineering should implement manifest generation, local sync state, and divergence comparison, then move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted this story from engineering intake to engineering active after document validation and lifecycle operations passed QA.
- `Why it matters`: It provides the local safety foundation for future Azure Blob `sync_pull` and `sync_push` without introducing remote mutation yet.
- `Acceptance criteria`: Existing criteria remain valid and testable. Azure Blob API calls remain out of scope.
- `Risks and assumptions`: Hashes should be the primary content-change signal. Move detection should use document id when available and degrade conservatively when ids are missing.
- `Completed work summary`: Refined and activated the sync manifest and local state story.
- `Next suggested or required step`: Engineering should implement local manifest/state/diff primitives and tests.
- `Next state recommendation`: engineering active
