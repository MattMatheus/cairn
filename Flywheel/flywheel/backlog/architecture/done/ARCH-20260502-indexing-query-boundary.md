# Architecture Story: Indexing Boundary And Query Contract ADR

## Metadata
- `id`: ARCH-20260502-indexing-query-boundary
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-indexing-query-boundary]
- `decision_owner`: Software Architect
- `success_metric`: A reviewer can implement Cairn local lookup and integrate CocoIndex without coupling Cairn core to derived artifact internals.

## Decision Scope
- Define the boundary between Cairn core indexing responsibilities and CocoIndex-derived context, including local SQLite metadata, local full-text lookup, semantic search, remote indexer endpoints, artifact ownership, refresh behavior, and query contracts.

## Problem Statement
- The north-star requires CocoIndex to power derived context without making Cairn core depend on every derived index being available. V1 needs a stable boundary so Cairn remains useful as a local markdown/MCP/sync tool while richer search can evolve behind a contract.

## Inputs
- Existing decisions:
  - `docs/product/north-star.md`
- Existing architecture artifacts:
  - `../cocoindex` reference repo outside the active Cairn repo.
- Constraints:
  - Cairn core is a self-contained Go CLI/MCP binary.
  - Indexing is optional but strongly recommended.
  - Cairn owns document discovery, frontmatter parsing, validation state, local metadata lookup, local full-text lookup, and MCP/CLI query contract.
  - CocoIndex owns semantic embeddings, richer summaries, entity extraction, graph features, and richer incremental processing.
  - Local SQLite metadata index lives at `/.cairn/index/cairn.db`.
  - Remote indexer should expose `/index/status`, `/index/refresh`, and `/search`.
  - Index artifacts are not normal document sync by default.
  - `sync_pull` should suggest an index refresh after new changes arrive.

## Outputs Required
- Decision updates:
  - `docs/adr/ADR-indexing-query-boundary.md` defining Cairn-owned indexes, CocoIndex-owned artifacts, local/remote search order, query request/response contracts, refresh semantics, and unavailable-index degradation.
- Architecture artifacts:
  - Ownership boundary table.
  - Query contract sketch for `search_context`.
  - Remote indexer endpoint contracts.
  - Search mode degradation examples.
  - Packaging/deployment options for local Docker/Podman and Azure Container Apps.
- Risks and tradeoffs:
  - Coupling to CocoIndex internals could slow Cairn iteration.
  - Too thin a query contract may hide useful retrieval features.
  - Optional indexers create uneven agent experiences across pods.
  - Remote indexer auth must align with Azure CLI/profile assumptions.

## Alternatives Considered
- Reimplement semantic indexing in Cairn core.
- Make CocoIndex required for all search.
- Treat derived index artifacts as normal workspace sync files.
- Expose CocoIndex artifact formats directly through MCP.

## Operational Impact
- Determines local-only usefulness.
- Determines remote indexer deployment and refresh workflows.
- Shapes performance, cost attribution, and isolation for pod-scoped search.

## Acceptance Criteria
1. ADR defines Cairn core versus CocoIndex ownership clearly.
2. ADR defines search mode order and graceful degradation behavior.
3. ADR defines remote indexer endpoint responsibilities at decision level.
4. ADR defines how index artifacts relate to document sync.
5. ADR identifies packaging and ACA auth questions to resolve or defer explicitly.

## Review Focus
- Confirm Cairn can operate without an indexer.
- Confirm semantic/richer context can evolve behind stable contracts.
- Confirm pod isolation and Azure profile assumptions are preserved.

## Next Step
- Architecture QA should review `docs/adr/ADR-indexing-query-boundary.md`.

## Architecture Handoff
- `Architecture decision`: Drafted `docs/adr/ADR-indexing-query-boundary.md` as a proposed ADR covering Cairn/CocoIndex ownership, local metadata/full-text search, remote indexer endpoints, search degradation, and index artifact sync boundaries.
- `Alternatives considered`: Reimplement semantic indexing in Cairn, require CocoIndex for all search, sync derived artifacts as normal documents, and expose CocoIndex artifacts directly.
- `Key risks`: Pods may have uneven search richness; query contracts must be stable without hiding useful retrieval features; ACA auth and packaging still need detail.
- `Follow-on implementation paths`: SQLite metadata index, local full-text, query schemas, CocoIndex pipeline prototype, local packaging, ACA deployment/auth, search degradation tests.
- `Next state recommendation`: Move to architecture QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: ADR was reviewed against the north-star document and the story acceptance criteria. It defines Cairn versus CocoIndex ownership, search mode ordering, graceful degradation, remote endpoint responsibilities, and the derived-artifact sync boundary.
- `Evidence quality call`: Sufficient for architecture acceptance; CocoIndex pipeline details, local packaging, and ACA auth enforcement remain explicit follow-on work.
- `Defects`: None.
- `Required fixes`: None.
- `Split decision`: No split required now. CocoIndex pipeline contracts and ACA deployment/auth may become child ADRs after reference exploration.
- `Completed work summary`: Accepted the indexing boundary and query contract ADR.
- `Next suggested or required step`: Use this ADR to seed implementation or exploration stories for SQLite metadata, full-text lookup, query schemas, CocoIndex prototype work, and ACA packaging/auth.
- `Next state recommendation`: Move to architecture done.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
