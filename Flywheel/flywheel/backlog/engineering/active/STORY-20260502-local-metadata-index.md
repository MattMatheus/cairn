# Story: Local Metadata Index Foundation

## Metadata
- `id`: STORY-20260502-local-metadata-index
- `owner_role`: Software Architect
- `status`: active
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary, ADR-document-model-lifecycle, ADR-mcp-operation-surface]
- `success_metric`: Cairn can create a local metadata index at `/.cairn/index/cairn.db`, populate it from managed markdown, and query stable result shapes without requiring CocoIndex.
- `release_scope`: required

## Problem Statement
- Cairn needs a local metadata lookup foundation before full-text search, semantic search, or remote indexer integration can be safely layered on top.

## Scope
- In:
  - Choose and add the minimal SQLite driver/dependency needed for local tests.
  - Define the local SQLite metadata schema for document path, id, title, slug, type, status, tags, actors, authors, source, and updated time.
  - Create the database at `/.cairn/index/cairn.db`.
  - Populate/update local metadata from managed markdown with valid frontmatter.
  - Implement local metadata lookup by title, slug, tag, status, type, path, actor, source, and recent changes where practical for this slice.
  - Return stable search/list result shapes aligned with the indexing ADR and `internal/mcpschema.SearchResult`.
  - Add tests for index creation, document indexing, metadata lookup, and stable result fields.
- Out:
  - Full-text lookup.
  - Semantic embeddings.
  - CocoIndex pipeline implementation.
  - Remote indexer deployment or calls.
  - MCP or CLI server wiring.

## Assumptions
- Valid document metadata parsing exists from the document model story.
- SQLite is acceptable for the local metadata index per the accepted ADR.
- Full-text lookup should follow as a separate story after metadata indexing is stable.

## Acceptance Criteria
1. Local metadata index can be created at `/.cairn/index/cairn.db`.
2. Managed markdown metadata can be indexed and queried locally.
3. Query results include path, title, type, status, slug, tags, updated time, score/match type, snippet, and provenance where available.
4. Metadata lookup supports at least title/query text, slug, tag, status, type, path, actor, source, and recent ordering.
5. Invalid or unmanaged markdown is skipped or reported without corrupting the index.
6. Tests cover database creation, indexing, representative metadata queries, and stable result shape.

## Validation
- Required checks:
  - Unit/integration tests for local index creation, update, and metadata query behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-indexing-query-boundary.md`.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`
- `STORY-20260502-mcp-schema-surface`

## Risks
- SQLite driver choice can affect portability; prefer a pure-Go driver unless there is a clear repo-local reason not to.
- Index freshness rules may need additional stories once sync is implemented.

## Open Questions
- Resolved for this slice: full-text search is split out and should not block the metadata index foundation.

## Next Step
- Engineering should implement the local metadata index package and tests, then move the story to engineering QA.

## PM Handoff
- `What changed`: Split the broad local index/query story and promoted the metadata index foundation into engineering active.
- `Why it matters`: Metadata lookup is the smallest durable index slice and gives future full-text, degradation, CLI, and MCP search work a stable base.
- `Acceptance criteria`: Focused on SQLite index creation, managed markdown metadata population, representative metadata queries, and stable result shapes.
- `Risks and assumptions`: SQLite dependency selection is part of this engineering slice. Full-text and search degradation remain separate follow-up work.
- `Completed work summary`: Created a bounded active story for local metadata indexing.
- `Next suggested or required step`: Engineering should implement local metadata index creation/population/query behavior with tests.
- `Next state recommendation`: engineering active
