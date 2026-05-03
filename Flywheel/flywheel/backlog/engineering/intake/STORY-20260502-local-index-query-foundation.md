# Story: Local Index And Query Foundation

## Metadata
- `id`: STORY-20260502-local-index-query-foundation
- `owner_role`: Software Architect
- `status`: intake
- `source`: direct
- `decision_refs`: [ADR-indexing-query-boundary, ADR-document-model-lifecycle]
- `success_metric`: Cairn can maintain local metadata/full-text lookup and return stable search result shapes without requiring CocoIndex.
- `release_scope`: required

## Problem Statement
- Cairn must remain useful when CocoIndex or a remote indexer is unavailable. Local metadata and full-text search are the foundation for that degraded-but-useful behavior.

## Scope
- In:
  - Define local SQLite metadata schema for document path, id, title, slug, type, status, tags, actors, authors, source, and updated time.
  - Populate/update local metadata from managed markdown.
  - Implement local metadata lookup.
  - Implement local full-text lookup or a clear v1-compatible abstraction for it.
  - Return stable search result shapes matching the indexing ADR.
  - Report unavailable semantic/remote modes with warnings and next steps.
  - Add tests for metadata lookup, full-text lookup, and degraded search response.
- Out:
  - CocoIndex pipeline implementation.
  - Remote indexer deployment.
  - Semantic embeddings.

## Assumptions
- Valid document metadata parsing exists from the document model story.
- SQLite is acceptable for the local metadata index per the accepted ADR.

## Acceptance Criteria
1. Local metadata index can be created at `/.cairn/index/cairn.db`.
2. Managed markdown metadata can be indexed and queried locally.
3. Full-text lookup works locally or has a tested abstraction ready for implementation.
4. Search results include path, title, type, status, slug, tags, updated time, score/match type, snippet, and provenance where available.
5. `auto` search degrades gracefully when semantic or remote index modes are unavailable.
6. Tests cover local metadata, full-text, and degradation behavior.

## Validation
- Required checks:
  - Unit/integration tests for local index creation, update, and query behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-indexing-query-boundary.md`.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`

## Risks
- Full-text implementation choice may vary by platform, especially Windows.
- Index freshness rules may need additional stories once sync is implemented.

## Open Questions
- Should local full-text use SQLite FTS immediately or hide the choice behind an interface?

## Next Step
- PM refinement should decide whether to split metadata indexing and full-text lookup into separate ready stories.
