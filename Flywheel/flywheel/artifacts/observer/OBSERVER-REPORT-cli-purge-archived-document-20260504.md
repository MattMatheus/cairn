# Observer Report: CLI-Only Purge Archived Document

## Summary
- Implemented `cairn purge` as a CLI-only hard-delete operation for archived documents.
- Required `--confirm-purge` before deletion.
- Enforced document-layer refusal for paths outside `archive/` and documents whose frontmatter status is not `archived`.
- Kept MCP purge/delete unavailable across read-only, local-writes, and remote-writes modes.

## QA
- `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/document ./internal/cli ./internal/mcpserver`
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote and complete `STORY-20260504-lifecycle-transition-enforcement`.
