# Story: MCP Lifecycle Mutation Adapters

## Metadata
- `id`: STORY-20260503-mcp-lifecycle-mutation-adapters
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface, ADR-document-model-lifecycle]
- `success_metric`: Capture, promote, and archive have transport-neutral MCP operation adapters that reuse document lifecycle primitives and response envelopes.
- `release_scope`: required

## Problem Statement
- Document lifecycle primitives exist, but MCP-facing mutation functions do not yet bind schema requests to those operations with consistent response envelopes.

## Scope
- In:
  - Implement local `capture_note`, `promote_document`, and `archive_document` adapters in `internal/mcpops`.
  - Reuse `internal/document.Workspace` lifecycle operations.
  - Return changed paths, durable document ids, warnings, provenance, and next steps.
  - Preserve MCP hard-delete/purge exclusion.
  - Add tests for successful mutations and validation failures.
- Out:
  - MCP server transport.
  - CLI command wiring.
  - Hard delete/purge.
  - Remote sync side effects.

## Assumptions
- Lifecycle operations are the source of truth for document mutation behavior.
- Adapter-level errors should preserve enough detail for MCP response mapping.

## Acceptance Criteria
1. Capture adapter writes under `/agents/{actor}` and returns changed path/id.
2. Promote adapter returns changed path/id and preserves lifecycle validation behavior.
3. Archive adapter returns archive path/id and does not expose purge.
4. All mutation responses use the common envelope and include next steps.
5. Tests cover success and representative failure cases.

## Validation
- Required checks:
  - Unit/integration tests for adapter behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against document lifecycle and MCP ADRs.

## Dependencies
- `STORY-20260502-capture-promotion-archive`
- `STORY-20260502-mcp-schema-surface`

## Risks
- Avoid duplicating lifecycle rules in the adapter.

## Open Questions
- None.

## Next Step
- PM should promote after progressive `read_document` passes QA.
