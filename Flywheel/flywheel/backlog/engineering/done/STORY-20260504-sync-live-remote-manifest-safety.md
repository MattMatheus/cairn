# Story: Sync Live Remote Manifest Safety

## Metadata
- `id`: STORY-20260504-sync-live-remote-manifest-safety
- `owner_role`: SRE
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-mcp-operation-surface]
- `success_metric`: Sync pull and push refuse stale or divergent live remote state before any local or remote mutation occurs.
- `release_scope`: required

## Problem Statement
- `SyncPush` and `SyncPull` build their safety plan from `.cairn/remote-manifest.json` rather than the configured remote store. In a real Azure profile, a missing or stale local manifest fixture can make remote changes invisible and allow push/pull mutations that should be refused.

## Scope
- In:
  - Read the live remote manifest through the configured `RemoteStore` before pull, push, and dry-run when remote sync is configured.
  - Compare the live remote manifest against local sync state and current local manifest.
  - Refuse mutation if the live remote manifest diverges from the last accepted base.
  - Preserve local state, local files, remote objects, and remote manifest on refusal.
  - Update CLI/MCP warnings and next steps for stale or unavailable live remote manifests.
- Out:
  - Automatic merge or conflict resolution.
  - New sync backend providers.
  - Remote auth redesign.

## Assumptions
- Local `.cairn/remote-manifest.json` may remain useful as a test fixture or local profile fallback, but configured remote profiles must prefer live remote state.

## Acceptance Criteria
1. `sync push` reads the live remote manifest before writing objects or publishing a manifest.
2. `sync pull` reads the live remote manifest before writing or removing local files.
3. Divergent live remote state returns a refused plan and does not mutate local state, local files, remote objects, or remote manifest.
4. Tests cover stale local fixture plus changed live remote manifest for both push and pull.
5. CLI and MCP responses include recovery guidance after stale or divergent remote state.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/syncstate ./internal/mcpops ./internal/cli`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Add fake remote-store tests proving `ReadManifest` is called before mutation methods.

## Dependencies
- None.

## Risks
- Live remote reads may introduce auth/network failure modes into dry-run and status flows.
- Care is needed to keep local-only profile behavior ergonomic.

## Open Questions
- Should local fixture fallback require an explicit local profile flag once remote sync is configured?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: Sync status, dry-run, pull, and push now load the live remote manifest from the configured remote store before planning mutations. Added coverage proving a stale local fixture cannot allow push to mutate divergent live remote state.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/syncstate ./internal/mcpops ./internal/cli`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm live manifest reads happen before writes and refused plans leave remote state untouched.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Targeted sync, MCP ops, CLI tests and full test suite passed. The added regression test verifies live remote divergence refuses push before write/manifest operations.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
