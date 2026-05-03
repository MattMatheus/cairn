# Story: Local Full-Text Search And Degradation

## Metadata
- `id`: STORY-20260503-local-full-text-search
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Cairn can run local full-text search and report graceful degradation when semantic or remote index modes are unavailable.
- `release_scope`: required

## Problem Statement
- After metadata indexing exists, Cairn needs local full-text lookup and predictable degraded search responses so agents can still retrieve useful context without CocoIndex or remote services.

## Scope
- In:
  - Implement deterministic local full-text lookup over managed markdown bodies, backed by the metadata index package for result metadata.
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
- For this slice, use a scanner-backed local full-text implementation instead of SQLite FTS. SQLite FTS can be introduced later as an optimization without changing the search response contract.

## Acceptance Criteria
1. Local full-text lookup works against managed markdown body content.
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
- Resolved for this slice: use a minimal scanner-backed implementation to keep behavior deterministic and portable.

## Next Step
- Engineering should implement local full-text and degraded `auto` search behavior, then move the story to engineering QA.

## PM Handoff
- `What changed`: Refined the story to use scanner-backed local full-text search for v1 rather than SQLite FTS.
- `Why it matters`: The local metadata index is complete, so full-text can now be layered on without adding platform-specific FTS behavior or remote dependencies.
- `Acceptance criteria`: Existing criteria are narrowed to managed markdown body search, `auto` ordering, stable search result shapes, and semantic/remote degradation reporting.
- `Risks and assumptions`: Scanner ranking/snippets should stay simple and deterministic. SQLite FTS can be revisited later as an implementation optimization.
- `Completed work summary`: Refined and activated the local full-text/degradation story.
- `Next suggested or required step`: Engineering should implement scanner-backed full-text search and degradation response tests.
- `Next state recommendation`: engineering active

## Engineering Handoff
- `What changed`: Added scanner-backed local full-text search over managed markdown bodies and a local search response layer that supports metadata, full-text, semantic-only degradation, and `auto` mode.
- `Why it matters`: Cairn can now return useful local metadata and body-content results while clearly reporting unavailable semantic and remote modes through the MCP response envelope.
- `Acceptance criteria`: Covered managed markdown body search, `auto` metadata-then-full-text ordering, stable `SearchResult` fields from the metadata index, semantic/remote degradation warnings, unavailable entries, provenance attempted modes, and next steps.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`
- `Risks and assumptions`: Scanner matching is intentionally simple and deterministic. SQLite FTS remains a later optimization if ranking or scale requires it.
- `QA focus areas`: Verify ADR ordering, degradation envelope fields, snippets/scores, and that full-text results are restricted to managed markdown with indexed metadata.
- `Completed work summary`: Implemented local full-text and degraded search behavior with tests.
- `Next suggested or required step`: QA should review the search behavior against the indexing and MCP ADRs, then move the story to done or file focused bugs.
- `Next state recommendation`: engineering qa

## QA Handoff
- `Verdict`: Pass.
- `Evidence summary`: Local full-text search scans managed markdown body content, uses indexed metadata for returned result fields, preserves stable `SearchResult` shape, and `auto` search attempts metadata then full-text before reporting semantic and remote unavailable modes through warnings, unavailable entries, provenance, and next steps.
- `Evidence quality call`: Strong for this scanner-backed v1 slice. Tests cover body search, full-text result fields, `auto` ordering, and semantic/remote degradation.
- `Defects`: None filed.
- `Required fixes`: None.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`; `bash Flywheel/flywheel/tools/validate_intake_items.sh`; `git diff --check`.
- `Completed work summary`: QA accepted the local full-text search and degradation story.
- `Next suggested or required step`: Close the cycle with an observer report and commit, then return to PM/planning for the next work item.
- `Next state recommendation`: engineering done
