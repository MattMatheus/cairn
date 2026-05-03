# Observer Report: MCP Read-Only Server Transport

## Cycle
- `id`: mcp-readonly-server-transport-20260503
- `story`: STORY-20260503-mcp-readonly-server-transport
- `completed_at`: 2026-05-03

## Result
- Added a minimal local read-only MCP transport.
- Story passed QA and moved to engineering done.

## Work Completed
- Added `internal/mcpserver`.
- Implemented standard-library line-delimited JSON-RPC handling for `initialize`, `tools/list`, and `tools/call`.
- Registered read-only tools: `get_bootstrap`, `search_context`, `list_documents`, `find_document`, `index_status`, and `read_document`.
- Bound tool handlers to `internal/mcpops.Local`.
- Added `cairn mcp readonly` for stdin/stdout serving.
- Added tests for tool registration, forbidden mutating tools, representative handler calls, JSON-RPC responses, and CLI launch.

## Verification
- `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/cli ./internal/mcpserver`
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
- Manual `go run ./cmd/cairn --root "$tmpdir" mcp readonly` smoke.

## Next Suggested Step
- Promote `STORY-20260503-cocoindex-contract-prototype` to define the external derived-context indexing contract.
