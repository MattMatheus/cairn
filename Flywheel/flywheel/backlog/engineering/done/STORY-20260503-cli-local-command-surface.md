# Story: CLI Local Command Surface

## Metadata
- `id`: STORY-20260503-cli-local-command-surface
- `owner_role`: Software Architect
- `status`: done
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
- Resolved for v1: use standard library flag parsing and keep command execution in an internal package for tests.

## Next Step
- Engineering should add the minimal local CLI command surface.

## PM Handoff
- Promoted on 2026-05-03 after init/config and validate workspace operations landed.
- Keep the CLI small and local-only; no sync mutations, remote auth, or MCP server work in this slice.
- Output must report completed work and suggested next steps for mutating commands.

## Engineering Handoff
- Added `cmd/cairn` entrypoint and testable `internal/cli.Run` command dispatcher.
- Wired local commands for `init`, `capture`, `promote`, `archive`, `validate`, `search`, and `index status`.
- Search refreshes the local metadata index before querying so fresh local captures are discoverable.
- Mutating commands print changed paths, document ids where available, and next suggested steps.
- Integration polish from CLI smoke:
  - validation now skips unmanaged markdown instead of warning on root pointer docs
  - generated onboarding docs now include starter frontmatter
  - warning-only validation output says “passed with warnings”
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/cli ./cmd/cairn`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" init ...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" capture ...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" validate`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" search ...`
  - `GOCACHE=/private/tmp/cairn-go-cache go run ./cmd/cairn --root "$tmpdir" index status`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed scoped commands are available through `go run ./cmd/cairn`.
- Confirmed command execution uses existing workspace, document lifecycle, local index, and MCP operation packages.
- Confirmed mutating command output includes changed paths, document ids where available, and next steps.
- Confirmed command tests cover parsing plus representative init, capture, promote, archive, validate, search, and index status flows.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`
  - Manual smoke across `init`, `capture`, `validate`, `search`, and `index status`.

## Next Suggested Step
- Promote sync status/conflict reporting so the CLI/MCP surfaces can expose local-vs-remote readiness before remote mutation commands.
