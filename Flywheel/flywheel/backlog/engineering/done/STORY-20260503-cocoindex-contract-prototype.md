# Story: CocoIndex Contract Prototype

## Metadata
- `id`: STORY-20260503-cocoindex-contract-prototype
- `owner_role`: Software Architect
- `status`: done
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
- Deferred: exact CocoIndex artifact formats and Azure Container Apps auth enforcement belong to packaging/deployment follow-up stories. Cairn v1 will depend only on HTTP status, refresh, and search contracts that map to Cairn schema shapes.

## Next Step
- Engineering should define the remote indexer contract and fake adapter tests.

## PM Handoff
- Promoted on 2026-05-03 after local MCP/read flows landed.
- Keep CocoIndex internals out of Cairn core; expose only stable status, refresh, and search contracts.
- Include reference repo review notes in the handoff and identify packaging/deployment follow-ups.

## Engineering Handoff
- Reviewed `../cocoindex` reference examples:
  - CocoIndex uses Python-native declarative flows over sources.
  - Examples chunk local files or session text, embed chunks, and store/query target databases such as Postgres/pgvector.
  - Target schemas vary by use case, so Cairn should avoid depending on CocoIndex artifact/table internals.
- Added `internal/remoteindex` with:
  - `Client` interface for `Status`, `Refresh`, and `Search`
  - HTTP client contract for `/index/status`, `/index/refresh`, and `/search`
  - `FakeClient` for contract tests
  - mapping from remote search/status responses into `mcpschema` shapes
- Added [cocoindex-contract-notes.md](/Users/foundry/CairnApp/cairn/docs/product/cocoindex-contract-notes.md).
- Follow-up stories identified:
  - local Docker/Podman packaging for the CocoIndex service
  - Azure Container Apps deployment and auth enforcement
  - refresh orchestration after sync pull/push
  - semantic search integration into `search_context`
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/remoteindex`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed the contract hides CocoIndex target table, embedding, and artifact internals.
- Confirmed `/index/status`, `/index/refresh`, and `/search` contract types/interfaces exist.
- Confirmed fake adapter tests map remote search/status into Cairn `mcpschema` shapes.
- Confirmed reference review notes and follow-up packaging/deployment stories are documented.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Return to PM backlog planning for the next implementation slice: likely packaging/deployment follow-ups or deeper sync pull/push orchestration.
