# Observer Report: MCP Remote Mutating Tools Gated Surface

## Result
- Accepted.

## Work Completed
- Added explicit remote-write MCP server option.
- Added `cairn mcp remote-writes`.
- Registered `sync_pull`, `sync_push`, and `index_refresh` only in remote-write mode.
- Preserved read-only and local lifecycle write mode boundaries.

## Verification
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Notes
- Default MCP server remains read-only.
- Local write mode remains lifecycle-only.
- Remote-write mode exposes sync/index mutations only.
- Delete and purge remain absent.

## Next Suggested Step
- Promote `STORY-20260504-remote-profile-config-client-wiring`.
