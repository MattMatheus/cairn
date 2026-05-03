# Story: CLI Local Command Surface

## Metadata
- `id`: STORY-20260503-cli-local-command-surface
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-document-model-lifecycle, ADR-indexing-query-boundary, ADR-sync-conflict-behavior]
- `success_metric`: Cairn has a minimal local CLI entrypoint for init, capture, promote, archive, validate, and search/status operations.
- `release_scope`: required

## Problem Statement
- Core operations exist as packages, but there is no user-facing CLI binary yet to exercise them outside tests.

## Scope
- In:
  - Add `cmd/cairn` entrypoint.
  - Wire local commands for `init`, `capture`, `promote`, `archive`, `validate`, `search`, and `index status`.
  - Use existing operation packages rather than duplicating logic.
  - Return concise human-readable output including completed work and suggested next steps.
  - Add command tests where practical.
- Out:
  - Remote sync mutation commands.
  - Interactive init.
  - MCP server.
  - Azure auth.

## Assumptions
- CLI can start with standard library flag parsing.

## Acceptance Criteria
1. `go run ./cmd/cairn ...` supports the scoped commands.
2. Commands call reusable operation packages.
3. Mutating command output includes changed paths/ids and next steps.
4. Tests cover at least command parsing and representative local commands.

## Validation
- Required checks:
  - Unit/integration tests for CLI command surface.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual smoke tests with `go run`.

## Dependencies
- `STORY-20260503-workspace-init-config`
- `STORY-20260503-validate-workspace-operation`
- `STORY-20260503-mcp-lifecycle-mutation-adapters`

## Risks
- Avoid large CLI framework churn before the command shape stabilizes.

## Open Questions
- Whether to use standard `flag` only or a small CLI package later.

## Next Step
- PM should promote after init/config and validation operations exist.
