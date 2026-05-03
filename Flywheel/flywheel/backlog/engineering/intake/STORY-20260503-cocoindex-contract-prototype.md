# Story: CocoIndex Contract Prototype

## Metadata
- `id`: STORY-20260503-cocoindex-contract-prototype
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary]
- `success_metric`: Cairn has a small prototype contract for exchanging search/index status data with a CocoIndex-backed indexer without exposing CocoIndex internals.
- `release_scope`: required

## Problem Statement
- The indexing ADR says CocoIndex should own richer derived context, but Cairn still needs a stable contract for status, refresh, and search responses before packaging or deployment work.

## Scope
- In:
  - Inspect the `cocoindex` reference repo for relevant pipeline and artifact patterns.
  - Define Cairn-side interface/types for `/index/status`, `/index/refresh`, and `/search` contracts.
  - Add fake adapter tests that map remote search responses into `mcpschema.SearchResult`.
  - Document assumptions and follow-up packaging/deployment stories.
- Out:
  - Running CocoIndex pipeline in Cairn.
  - Docker/Podman packaging.
  - Azure Container Apps deployment.
  - Semantic embedding implementation.

## Assumptions
- Contract should hide CocoIndex artifact details behind Cairn result shapes.

## Acceptance Criteria
1. Reference repo review is summarized in the story handoff or a small doc.
2. Remote indexer contract types/interfaces exist.
3. Fake adapter maps remote search/status into Cairn response shapes.
4. Tests cover status and search mapping.
5. Follow-up packaging/deployment stories are identified.

## Validation
- Required checks:
  - Unit tests for contract mapping.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against indexing ADR.

## Dependencies
- `STORY-20260503-local-full-text-search`
- `STORY-20260502-mcp-schema-surface`

## Risks
- Avoid coupling to unstable CocoIndex internals.

## Open Questions
- Exact remote artifact format and auth enforcement details.

## Next Step
- PM should promote after core local MCP/read flows are stable.
