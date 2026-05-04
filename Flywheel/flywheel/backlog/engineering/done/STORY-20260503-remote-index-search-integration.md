# Story: Remote Index Search Integration

## Metadata
- `id`: STORY-20260503-remote-index-search-integration
- `owner_role`: Software Architect
- `status`: done
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

## Engineering Handoff
- Implemented 2026-05-03.
- Added optional `remoteindex.Client` integration behind `localindex.Search`.
- Auto mode keeps local metadata and full-text ordering, then attempts configured semantic remote search.
- Semantic mode calls the configured remote indexer directly.
- Remote results map through the existing stable `mcpschema.SearchResult` shape.
- Remote failures degrade into warnings, unavailable modes, and next steps instead of failing local search.
- Wired `mcpops.Local.SearchContext` to pass the configured remote index client.

## QA Handoff
- Accepted 2026-05-03.
- Verified auto mode attempts metadata, full-text, semantic in order.
- Verified semantic mode calls configured remote indexer.
- Verified merged local/remote results dedupe by path and keep local results first.
- Verified remote result provenance and match type are preserved in Cairn schema.
- Verified unconfigured and failing remote indexers degrade gracefully.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260503-index-refresh-contract-wiring`.
