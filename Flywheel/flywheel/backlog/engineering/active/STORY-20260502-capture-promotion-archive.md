# Story: Capture Promotion And Archive Lifecycle Operations

## Metadata
- `id`: STORY-20260502-capture-promotion-archive
- `owner_role`: Software Architect
- `status`: active
- `source`: direct
- `decision_refs`: [ADR-document-model-lifecycle, ADR-mcp-operation-surface]
- `success_metric`: Cairn can create, promote, and archive managed documents while preserving durable identity and lifecycle rules.
- `release_scope`: required

## Problem Statement
- Agents and humans need first-class lifecycle operations instead of ad hoc file writes so canonical knowledge, ADR numbering, and archive behavior stay reliable.

## Scope
- In:
  - Implement capture operation for agent-authored notes under `/agents/{actor}/...`.
  - Generate durable Cairn document ids.
  - Implement promotion to `proposed` and `canonical`.
  - Preserve durable metadata during promotion.
  - Move promoted documents to configured destination folders.
  - Assign ADR numbers when `decision` documents become `canonical`.
  - Implement archive operation preserving original path under `/archive`.
  - Add tests for identity preservation, status transitions, promotion moves, ADR numbering, and archive moves.
- Out:
  - Hard delete or purge implementation.
  - Azure sync behavior.
  - MCP server wiring beyond exposing reusable operation functions.

## Assumptions
- Frontmatter parsing and validation are available from `STORY-20260502-document-frontmatter-validation`.
- ADR number allocation can start with local deterministic behavior and later be guarded by sync conflict handling.
- This story should extend the existing `internal/document` package or add the smallest adjacent package needed for filesystem lifecycle operations.
- MCP and CLI adapters should remain out of scope; expose reusable operation functions that later adapters can call.

## Acceptance Criteria
1. Capture creates a managed markdown document with valid core frontmatter.
2. Promotion to `proposed` repairs or adds required frontmatter when possible.
3. Promotion to `canonical` blocks until core frontmatter is valid.
4. Canonical decision promotion assigns an ADR number and final `/decisions/ADR-000N-slug.md` path.
5. Archive sets `status: archived` and moves the document under `/archive` preserving original path.
6. Tests cover capture, proposed promotion, canonical promotion, ADR numbering, and archive behavior.

## Validation
- Required checks:
  - Relevant unit/integration tests for lifecycle operations.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-document-model-lifecycle.md`.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`

## Risks
- ADR number allocation may need extra locking once sync is implemented.
- Promotion moves can surprise users if paths are not reported clearly.

## Open Questions
- Resolved for this slice: CLI-only purge is deferred until after archive is working and should be created as a separate story when needed.

## Next Step
- Engineering should implement capture, promotion, ADR numbering, and archive lifecycle operations, then move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted this story from engineering intake to engineering active as the next dependency after frontmatter validation.
- `Why it matters`: It turns the accepted document lifecycle ADR into reusable product operations for later CLI and MCP adapters.
- `Acceptance criteria`: Existing criteria remain valid and testable. Purge remains out of scope.
- `Risks and assumptions`: ADR number allocation may be local-only in this slice and later protected by sync conflict behavior. File moves should report paths clearly in operation results.
- `Completed work summary`: Refined and activated the capture, promotion, and archive lifecycle story.
- `Next suggested or required step`: Engineering should implement reusable lifecycle operations on top of `internal/document`.
- `Next state recommendation`: engineering active
