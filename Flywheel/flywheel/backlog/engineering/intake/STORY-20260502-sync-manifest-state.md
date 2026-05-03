# Story: Sync Manifest And Local State Foundation

## Metadata
- `id`: STORY-20260502-sync-manifest-state
- `owner_role`: Software Architect
- `status`: intake
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
- Remote manifest loading can be represented by local fixtures in this story.

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
- Should local sync state store the full base manifest or a compact normalized form?

## Next Step
- PM refinement should decide whether to split `.cairnignore` parsing if it is nontrivial in the chosen implementation stack.
