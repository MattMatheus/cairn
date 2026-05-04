# Observer Report: MCP Mutating Tools Gated Surface

## Result
- Accepted.

## Work Completed
- Added explicit write-enabled MCP server option.
- Preserved read-only default server behavior.
- Registered `capture_note`, `promote_document`, and `archive_document` only in write-enabled mode.
- Wired mutation handlers to existing lifecycle operation adapters.
- Added CLI entry `cairn mcp local-writes`.

## Verification
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Notes
- Default MCP server remains read-only.
- Write-enabled mode exposes only local lifecycle mutations.
- Sync mutations, index refresh, delete, and purge remain absent.
- Representative capture mutation returns common envelope and writes the document.

## Next Suggested Step
- Promote `STORY-20260503-cocoindex-local-packaging`.
