# Story: MCP Read-Only Server Transport

## Metadata
- `id`: STORY-20260503-mcp-readonly-server-transport
- `owner_role`: Software Architect
- `status`: done
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
- Resolved for v1 foundation: implement a small isolated JSON-RPC/MCP transport package using the standard library, avoiding dependency churn until the protocol surface stabilizes.

## Next Step
- Engineering should implement the read-only MCP transport and tests.

## PM Handoff
- Promoted on 2026-05-03 after read-only operation adapters and CLI coverage landed.
- Keep the server read-only in this slice; do not expose capture, promote, archive, sync pull/push, index refresh, purge, or delete.
- Preserve envelope responses from `internal/mcpops`.

## Engineering Handoff
- Added `internal/mcpserver` with a small standard-library JSON-RPC/MCP transport.
- Registered read-only tools only: `get_bootstrap`, `search_context`, `list_documents`, `find_document`, `index_status`, and `read_document`.
- Bound handlers directly to `internal/mcpops.Local` operation adapters.
- Preserved common envelope shape inside MCP tool text responses.
- Added `cairn mcp readonly` to serve line-delimited JSON-RPC over stdin/stdout.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/cli ./internal/mcpserver`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
  - `printf ... | GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" mcp readonly`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed scoped read-only tools are registered.
- Confirmed handlers call `internal/mcpops.Local` operation adapters.
- Confirmed MCP tool calls preserve the common envelope shape inside JSON text content.
- Confirmed mutating tools, sync mutation tools, index refresh, purge, and delete are absent.
- Confirmed `cairn mcp readonly` launches the stdin/stdout transport.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote the CocoIndex contract prototype story to define the richer derived-context boundary behind Cairn's stable search contract.
