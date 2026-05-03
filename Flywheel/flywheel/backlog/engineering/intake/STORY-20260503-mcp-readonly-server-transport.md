# Story: MCP Read-Only Server Transport

## Metadata
- `id`: STORY-20260503-mcp-readonly-server-transport
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface]
- `success_metric`: Cairn can expose the local read/search MCP operation subset through an MCP server transport.
- `release_scope`: required

## Problem Statement
- MCP operation schemas and local operation adapters exist, but agents cannot call them through an actual MCP server yet.

## Scope
- In:
  - Choose minimal Go MCP server library or protocol implementation.
  - Expose read-only tools: `get_bootstrap`, `search_context`, `list_documents`, `find_document`, `index_status`, and `read_document` when available.
  - Bind schema request/response types to operation adapters.
  - Add tests for tool registration and representative handler responses.
- Out:
  - Mutating MCP tools.
  - Remote profile behavior.
  - Hard delete/purge.

## Assumptions
- Read-only transport should land before mutating MCP tools for safety.

## Acceptance Criteria
1. Server registers scoped read-only tools.
2. Handlers call `internal/mcpops` rather than duplicating behavior.
3. Tool responses preserve common envelope shape.
4. Purge/delete is absent.
5. Tests cover registration and representative calls.

## Validation
- Required checks:
  - Unit/integration tests for server registration/handlers.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against MCP ADR.

## Dependencies
- `STORY-20260503-mcp-read-document-progressive`
- `STORY-20260503-mcp-local-read-search-operations`

## Risks
- MCP library choice could create churn; keep boundary isolated.

## Open Questions
- Which Go MCP server implementation should be used.

## Next Step
- PM should refine once read-only adapter coverage is complete.
