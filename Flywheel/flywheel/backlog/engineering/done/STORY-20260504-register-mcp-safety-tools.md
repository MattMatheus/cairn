# Story: Register MCP Safety Tools

## Metadata
- `id`: STORY-20260504-register-mcp-safety-tools
- `owner_role`: Software Architect
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-mcp-operation-surface, ADR-sync-conflict-behavior]
- `success_metric`: MCP clients can call the v1 `validate_workspace` and `sync_status` tools before attempting workspace mutations.
- `release_scope`: required

## Problem Statement
- The v1 schema and ADR include `validate_workspace` and `sync_status`, but the MCP server does not register them. Remote-write mode exposes pull/push without the status inspection tool agents need for safe operation.

## Scope
- In:
  - Register `validate_workspace` in read-only MCP mode.
  - Register `sync_status` in read-only and remote-write MCP modes.
  - Add input schemas for validation and sync status requests.
  - Ensure existing write gating still excludes local mutations and remote mutations in the proper modes.
  - Add server and JSON-RPC tests for callable safety tools.
- Out:
  - New MCP tools beyond the v1 schema.
  - Changes to purge/delete exposure.

## Assumptions
- `sync_status` is read-only and safe to expose in all MCP modes.

## Acceptance Criteria
1. `cairn mcp readonly` lists and can call `validate_workspace`.
2. `cairn mcp readonly` lists and can call `sync_status`.
3. `cairn mcp remote-writes` lists `sync_status`, `sync_pull`, `sync_push`, and `index_refresh`.
4. Local-write and remote-write permission boundaries remain covered by tests.
5. No purge or hard delete tool is exposed.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/mcpserver ./internal/mcpops ./internal/mcpschema`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Exercise `tools/list` and `tools/call` through JSON-RPC test cases.

## Dependencies
- None.

## Risks
- Tool registration order tests may need updates without loosening permission assertions.

## Open Questions
- Should `index_status` also be listed in remote-write mode documentation as a safety preflight?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: MCP read-only tools now include `validate_workspace` and `sync_status`, so all MCP modes inherit read-only safety preflight tools while write gates remain intact.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/mcpserver ./internal/mcpops ./internal/mcpschema`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm no purge/delete tool exposure and local/remote write boundaries remain unchanged.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Server registration tests were updated and targeted MCP tests plus full suite passed.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
