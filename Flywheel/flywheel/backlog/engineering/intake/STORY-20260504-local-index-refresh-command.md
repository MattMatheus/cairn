# Story: Local Index Refresh Command

## Metadata
- `id`: STORY-20260504-local-index-refresh-command
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Users and agents can refresh Cairn's local metadata index through the same operation surface as remote index refresh.
- `release_scope`: required

## Problem Statement
- Cairn can index workspaces internally, but `index_refresh` currently represents remote refresh behavior and `cairn index refresh` does not rebuild the local metadata index.

## Scope
- In:
  - Add local metadata index refresh behavior to `IndexRefresh`.
  - Make `cairn index refresh` rebuild the local index when no remote indexer is configured.
  - Report local changed paths and next steps.
  - Preserve remote refresh behavior when a remote indexer is configured.
  - Add tests for local-only and remote-configured refresh behavior.
- Out:
  - CocoIndex local embedding refresh.
  - Scheduled refresh.
  - Watch mode.

## Assumptions
- Local refresh should be synchronous and deterministic.

## Acceptance Criteria
1. Local-only `index_refresh` rebuilds `.cairn/index/cairn.db`.
2. Response reports `local_refreshed: true` and changed path.
3. Remote-configured refresh still reports accepted/refreshed remote state.
4. CLI output distinguishes local and remote refresh.
5. Tests cover local-only and remote-configured paths.

## Validation
- Required checks:
  - Local index and mcpops tests.
  - CLI tests.

## Dependencies
- `STORY-20260502-local-metadata-index`
- `STORY-20260503-index-refresh-contract-wiring`

## Risks
- Keep local index refresh independent from remote semantic refresh.

## Open Questions
- Whether local refresh should also run automatically after capture/promote/archive.

## Next Step
- Engineering should implement early in the next batch because it improves day-to-day usability.
