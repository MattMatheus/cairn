# Story: MCP Tool Schema Surface

## Metadata
- `id`: STORY-20260502-mcp-schema-surface
- `owner_role`: Software Architect
- `status`: intake
- `source`: direct
- `decision_refs`: [ADR-mcp-operation-surface, ADR-document-model-lifecycle, ADR-sync-conflict-behavior, ADR-indexing-query-boundary]
- `success_metric`: The v1 MCP tool surface has concrete request and response schemas with shared warnings, provenance, and next-step reporting.
- `release_scope`: required

## Problem Statement
- The accepted MCP ADR defines the product operation surface at decision level. Implementation needs concrete schemas before server wiring can be built safely.

## Scope
- In:
  - Define request/response schemas for v1 MCP tools.
  - Define common response envelope with `ok`, `data`, `warnings`, `unavailable`, `next_steps`, and `provenance`.
  - Define actor and profile fields.
  - Define read modes and search modes.
  - Define error/warning shapes for validation, sync, and index degradation.
  - Add schema tests or golden examples.
- Out:
  - Full MCP server transport implementation.
  - Concrete lifecycle, sync, or indexing operation internals beyond schema integration points.
  - Hard delete/purge exposure.

## Assumptions
- Operation internals can be implemented behind these schemas in later stories.
- Schemas should support CLI parity where practical.

## Acceptance Criteria
1. Every v1 MCP tool from the ADR has a concrete request and response schema.
2. All responses include a consistent way to report warnings and next steps.
3. Mutating tools include changed paths and durable ids where applicable.
4. Read/search schemas support progressive disclosure.
5. Hard delete/purge is absent from the schema surface.
6. Tests or examples validate representative schema payloads.

## Validation
- Required checks:
  - Schema validation tests or golden example checks.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-mcp-operation-surface.md`.

## Dependencies
- Accepted ADRs.
- May proceed before full operation implementations if schemas are isolated.

## Risks
- Overly detailed schemas may churn before implementation learns enough.
- Under-specified schemas may allow unsafe agent behavior.

## Open Questions
- Should schemas live in code, generated docs, or both?

## Next Step
- PM refinement should decide whether this runs before or alongside lifecycle operation implementation.
