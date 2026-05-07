# Architecture Story: Core V1 Indexing Boundary Refresh

## Metadata
- `id`: ARCH-20260507-core-v1-indexing-boundary-refresh
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary]
- `decision_owner`: Software Architect
- `success_metric`: A new engineer can identify local SQLite metadata/full-text search as the v1 retrieval path and CocoIndex as deferred rich retrieval.

## Decision Scope
- Refresh the indexing and query boundary for Cairn Core v1 so local search is the required retrieval surface and remote semantic indexing is explicitly deferred or optional.

## Problem Statement
- The current direction risks making CocoIndex, pgvector, and a remote indexer feel like required v1 infrastructure. Cairn v1 should instead prove the local-first workspace, validation, local index/search, MCP, blob sync, and conflict handling loop before adding rich retrieval services.

## Inputs
- Existing decisions:
  - `docs/adr/ADR-indexing-query-boundary.md`
  - `docs/adr/ADR-sync-conflict-behavior.md`
  - `docs/adr/ADR-mcp-operation-surface.md`
  - `docs/adr/ADR-local-development-emulation.md`
- Existing architecture artifacts:
  - `Flywheel/flywheel/artifacts/planning/PLAN-20260507-cairn-core-v1-rescope.md`
  - `Flywheel/flywheel/artifacts/planning/PLAN-20260506-local-dev-cocoindex-environment.md`
- Constraints:
  - Cairn core remains a self-contained Go CLI/MCP binary.
  - Engineers must be able to see and edit local markdown files.
  - V1 retrieval must not require CocoIndex, pgvector, Postgres, or a remote indexer.
  - The MCP contract should remain stable enough for deferred semantic search to return later.

## Outputs Required
- Decision updates:
  - Update `docs/adr/ADR-indexing-query-boundary.md` or create a superseding ADR section that clearly states the Cairn Core v1 boundary.
- Architecture artifacts:
  - V1 ownership boundary table.
  - Search mode behavior for `metadata`, `full_text`, `auto`, and deferred `semantic`.
  - Index refresh semantics where v1 refresh rebuilds local SQLite only.
  - Explicit deferred-work list for CocoIndex, remote indexer, pgvector, and production remote index auth.
- Risks and tradeoffs:
  - Reduced remote retrieval richness in exchange for simpler adoption.
  - Potential future migration path for optional rich retrieval adapters.
  - Avoiding MCP contract churn while changing v1 implementation expectations.

## Alternatives Considered
- Keep CocoIndex integration in the v1 mainline.
- Remove indexing concepts entirely from v1.
- Make blob storage canonical and drop local-first sync.
- Split the MCP contract into separate core and rich-retrieval versions immediately.

## Operational Impact
- Defines what a new engineer must run locally.
- Determines which services are required in the v1 smoke path.
- Clarifies whether remote indexer failures are product blockers or optional degradation.

## Acceptance Criteria
1. Decision text states that local SQLite metadata/full-text search is required for v1.
2. Decision text states that CocoIndex, remote semantic search, pgvector, and remote indexer deployment are deferred or optional for v1.
3. `index_refresh` and `search_context` v1 semantics are explicit.
4. Blob sync and conflict handling remain in the v1 core boundary.
5. Follow-on engineering work is clearly split and testable.

## Review Focus
- Confirm the v1 boundary is small enough for first external engineer adoption.
- Confirm deferred rich retrieval can return later without breaking the MCP surface.
- Confirm local-first visibility and blob sync remain central.

## Next Step
- PM should refine the Cairn Core v1 engineering intake stories and activate the first implementation/documentation batch.

## Architecture Handoff
- `Architecture decision`: Refreshed `docs/adr/ADR-indexing-query-boundary.md` so Cairn Core v1 requires local SQLite metadata/full-text retrieval, keeps blob sync and conflict refusal in core, and treats CocoIndex, pgvector, remote semantic search, remote indexer deployment, and production remote-index auth as deferred rich-retrieval work.
- `Alternatives considered`: Kept CocoIndex in v1 mainline, removed indexing entirely, made blob storage canonical instead of local-first sync, split MCP into core/rich versions immediately, and synced generated index artifacts as normal documents.
- `Key risks`: V1 search is less rich than the prior ambition; stale docs may still imply the remote indexer is required; optional rich retrieval must return later without changing the stable MCP result shape.
- `Follow-on implementation paths`: Update docs/quickstart, verify local-only index refresh and search behavior, add local blob-sync smoke without service dependencies, and relabel CocoIndex artifacts as deferred/reference where needed.
- `Next state recommendation`: Move to architecture QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed `docs/adr/ADR-indexing-query-boundary.md` against all five acceptance criteria. The ADR explicitly requires local SQLite metadata/full-text retrieval for v1, defers CocoIndex, remote semantic search, pgvector, remote indexer deployment, and production remote-index auth, defines `index_refresh` and `search_context` v1 behavior, keeps blob sync/conflict refusal in core, and lists concrete follow-on engineering work.
- `Evidence quality call`: Sufficient for architecture acceptance. This was a decision-artifact QA pass, so no code tests were required.
- `Defects`: None.
- `Required fixes`: None.
- `Residual risks`: Product/user docs and local-dev artifacts may still contain stale remote-indexer-first language; this is covered by follow-on engineering intake.
- `Next state recommendation`: Move to architecture done.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
