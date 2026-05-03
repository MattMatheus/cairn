# Story: Local Full-Text Search And Degradation

## Metadata
- `id`: STORY-20260503-local-full-text-search
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Cairn can run local full-text search and report graceful degradation when semantic or remote index modes are unavailable.
- `release_scope`: required

## Problem Statement
- After metadata indexing exists, Cairn needs local full-text lookup and predictable degraded search responses so agents can still retrieve useful context without CocoIndex or remote services.

## Scope
- In:
  - Implement local full-text lookup or a v1-compatible abstraction backed by the metadata index package.
  - Return stable search result shapes matching the indexing ADR and MCP schema.
  - Implement `auto` search ordering across metadata and full-text local modes.
  - Report unavailable semantic/remote modes with warnings, unavailable entries, provenance, and next steps.
  - Add tests for full-text lookup and degraded search response.
- Out:
  - CocoIndex pipeline implementation.
  - Remote indexer deployment or calls.
  - Semantic embeddings.
  - MCP or CLI server transport wiring.

## Assumptions
- Local metadata index foundation is complete.
- SQLite FTS can be used if the selected SQLite dependency supports it; otherwise hide the choice behind an interface and keep the behavior testable.

## Acceptance Criteria
1. Local full-text lookup works or has a tested abstraction ready for the first implementation.
2. `auto` search combines/delegates metadata and full-text local modes in ADR order.
3. Search results include path, title, type, status, slug, tags, updated time, score/match type, snippet, and provenance where available.
4. Semantic and remote modes degrade gracefully when unavailable.
5. Tests cover local full-text behavior and degraded search response.

## Validation
- Required checks:
  - Unit/integration tests for full-text and degradation behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-indexing-query-boundary.md` and `docs/adr/ADR-mcp-operation-surface.md`.

## Dependencies
- `STORY-20260502-local-metadata-index`

## Risks
- SQLite FTS availability may vary by driver/platform.
- Ranking and snippets can churn; keep v1 behavior simple and deterministic.

## Open Questions
- Should local full-text use SQLite FTS immediately or a minimal scanner behind an interface?

## Next Step
- PM should refine after metadata indexing passes QA.
