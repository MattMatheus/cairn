# Story: Remote Index Search Integration

## Metadata
- `id`: STORY-20260503-remote-index-search-integration
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: `search_context` can optionally call the remote indexer contract and merge/degrade results without exposing remote internals.
- `release_scope`: required

## Problem Statement
- Cairn has a remote indexer contract but local search still reports semantic/remote modes as unavailable.

## Scope
- In:
  - Add optional remote index client integration behind `search_context`.
  - Preserve local metadata and full-text fallback behavior.
  - Map remote semantic results into stable `mcpschema.SearchResult`.
  - Surface warnings/unavailable modes when remote indexer is not configured or unavailable.
  - Add tests for remote available, remote unavailable, and merged local/remote results.
- Out:
  - Running CocoIndex pipelines.
  - Remote index refresh orchestration.
  - Azure Container Apps deployment/auth enforcement.

## Assumptions
- Remote indexer config can remain minimal and injectable for tests.

## Acceptance Criteria
1. `search_context` auto mode attempts configured remote semantic search after local modes.
2. Semantic mode calls remote indexer when configured.
3. Remote results preserve Cairn search result shape and provenance.
4. Remote unavailable responses degrade gracefully with warnings and next steps.
5. Tests cover available and unavailable remote indexer behavior.

## Validation
- Required checks:
  - Unit tests for search integration.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against indexing ADR search order.

## Dependencies
- `STORY-20260503-cocoindex-contract-prototype`

## Risks
- Avoid making remote search required for local usability.

## Open Questions
- Exact config source for remote index endpoint and auth token provider in v1.

## Next Step
- Engineering should wire remote search after sync dry-run planning or when indexer config is ready.

