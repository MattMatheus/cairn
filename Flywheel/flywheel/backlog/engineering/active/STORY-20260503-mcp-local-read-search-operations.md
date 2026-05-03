# Story: MCP Local Read And Search Operations

## Metadata
- `id`: STORY-20260503-mcp-local-read-search-operations
- `owner_role`: Software Architect
- `status`: active
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface, ADR-indexing-query-boundary, ADR-document-model-lifecycle]
- `success_metric`: Cairn has reusable local operation implementations for the read/search subset of the MCP surface without requiring MCP server transport.
- `release_scope`: required

## Problem Statement
- Cairn now has MCP schemas, document parsing, lifecycle primitives, and local search primitives, but the MCP-facing read/search operations do not yet have reusable implementation functions behind the schema contract.

## Scope
- In:
  - Add an isolated local MCP operation adapter package that consumes `internal/mcpschema` request types and returns `mcpschema.Envelope[...]` responses.
  - Implement local-profile behavior for `search_context`, `list_documents`, `find_document`, `index_status`, and a minimal `get_bootstrap`.
  - Reuse `internal/localindex` for metadata/full-text search and result shapes.
  - Reuse `internal/document` parsing/validation where needed.
  - Include warnings, unavailable modes, provenance, and next steps consistently with the MCP schema story.
  - Add tests for representative request/response behavior and local-only degradation.
- Out:
  - MCP server transport.
  - Remote `pod-remote` calls.
  - Mutating MCP operations such as capture, promote, archive, sync pull/push, or index refresh.
  - Full `read_document` progressive section parsing beyond a follow-up story.

## Assumptions
- Local metadata and full-text search foundations are complete.
- The adapter should be transport-neutral so a later MCP server can call it directly.
- `read_document` deserves its own follow-up because progressive document section parsing is a larger behavior slice.

## Acceptance Criteria
1. `search_context` local mode returns metadata/full-text results and degradation details through the common envelope.
2. `list_documents` returns stable document summaries from the local metadata index.
3. `find_document` can locate documents by id, slug, title, path, type, status, or tag using local metadata.
4. `index_status` reports local index availability/freshness and remote unavailable state.
5. `get_bootstrap` returns compact workspace context and useful next steps without reading large document bodies.
6. Tests cover local success responses, unavailable/degraded remote behavior, and response envelope fields.

## Validation
- Required checks:
  - Unit/integration tests for local MCP operation adapter behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-mcp-operation-surface.md` and `docs/adr/ADR-indexing-query-boundary.md`.

## Dependencies
- `STORY-20260502-mcp-schema-surface`
- `STORY-20260502-local-metadata-index`
- `STORY-20260503-local-full-text-search`

## Risks
- It is easy to accidentally build MCP transport or remote behavior in this slice; keep the package transport-neutral and local-only.
- `find_document` and `list_documents` may need richer filters later, but should start from the schema fields already defined.

## Open Questions
- Resolved for this slice: leave `read_document`, mutating lifecycle tools, sync mutation tools, and MCP server transport to later stories.

## Next Step
- Engineering should implement the local read/search operation adapter and tests, then move the story to engineering QA.

## PM Handoff
- `What changed`: Created and promoted the next engineering story after local search foundations passed QA.
- `Why it matters`: It turns the schema and local search primitives into reusable product-operation functions that future MCP transport can call.
- `Acceptance criteria`: Focused on local read/search operations with envelopes, degradation reporting, and no transport or mutation behavior.
- `Risks and assumptions`: Keep this local-only and transport-neutral. Progressive `read_document` and mutating MCP operations should remain separate.
- `Completed work summary`: Added a bounded active story for local MCP read/search operations.
- `Next suggested or required step`: Engineering should implement the local adapter package and tests.
- `Next state recommendation`: engineering active
