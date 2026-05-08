# Story: MCP large-request handling and YAML-safe frontmatter

## Metadata
- `id`: STORY-20260508-mcp-large-request-and-yaml-escape
- `owner_role`: Software Architect
- `status`: ready
- `source`: planning
- `decision_refs`: [ARCH-20260502-mcp-operation-surface, ARCH-20260502-document-model-lifecycle]
- `success_metric`: MCP requests with bodies up to 16 MB succeed; frontmatter round-trips losslessly for titles containing YAML-special characters.
- `release_scope`: required

## Problem Statement
- `internal/mcpserver/server.go:99-122` uses `bufio.Scanner` with the 64 KB default buffer; large `capture_note` bodies cause silent truncation and server exit (High, H2).
- `internal/document/lifecycle.go:406-431` writes raw values into YAML frontmatter; titles or fields containing `:`, leading `-`, `#`, `[`, `&`, or newlines produce invalid YAML and silent corruption on round-trip (High, H6).

## Scope
- In:
  - Replace the MCP scanner with a streaming JSON-RPC decoder, or expand the scanner buffer to 16 MB.
  - YAML-escape string fields in `renderDocument` (use `strconv.Quote` style or a minimal YAML quoter consistent with `quoteRepoValue` in `repos.go:247`).
  - Round-trip parser test for special characters.
- Out:
  - Full YAML library swap.
  - MCP transport rework.

## Assumptions
- 16 MB is a sufficient hard cap for any single MCP frame in v1.
- Existing markdown corpus does not contain titles needing re-escape on read; if it does, parser is tolerant.

## Acceptance Criteria
1. MCP server handles a `capture_note` with a 1 MB body without truncation; integration test asserts response.
2. Frontmatter round-trip test covers titles containing `:`, leading `-`, `#`, `[`, `&`, `"`, and embedded newlines; output re-parses to identical metadata.
3. Existing frontmatter fixtures continue to parse.

## Validation
- Required checks:
  - `go test ./internal/mcpserver/...`
  - `go test ./internal/document/...`
  - `make pilot-check`
- Additional checks:
  - Manual: `cairn note` with a long pasted body via MCP and via CLI.

## Dependencies
- None.

## Risks
- Choosing `json.NewDecoder` changes framing semantics; verify MCP newline-delimited contract still holds.

## Open Questions
- Should we always quote string fields, or only when special characters detected? Prefer minimal-quoting for diff-friendliness.

## Next Step
- PM ranks; engineering implements.
