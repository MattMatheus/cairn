# Story: Index Refresh Contract Wiring

## Metadata
- `id`: STORY-20260503-index-refresh-contract-wiring
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Cairn can request remote index refresh through the stable indexer contract and report accepted/refreshed state.
- `release_scope`: required

## Problem Statement
- The remote indexer contract includes `/index/refresh`, but Cairn does not expose or consume refresh responses yet.

## Scope
- In:
  - Add operation adapter for index refresh using `remoteindex.Client`.
  - Preserve local-only graceful degradation.
  - Return changed-path/result shape appropriate for MCP/CLI.
  - Suggest refresh after successful sync pull where hooks exist.
  - Add tests with fake remote index client.
- Out:
  - Running CocoIndex locally.
  - Azure Container Apps deployment.
  - Scheduled/automatic refresh.

## Assumptions
- Refresh request can be accepted asynchronously by the remote service.

## Acceptance Criteria
1. Refresh reports accepted/refreshed/job id where provided.
2. Missing remote indexer degrades with warnings and next steps.
3. Responses preserve common envelope shape.
4. Tests cover accepted, refreshed, and unavailable cases.

## Validation
- Required checks:
  - Unit tests for adapter behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against indexing ADR endpoint responsibilities.

## Dependencies
- `STORY-20260503-cocoindex-contract-prototype`

## Risks
- Avoid implying refresh is synchronous when the remote service accepts jobs.

## Open Questions
- Whether refresh should be CLI-only until remote deployment exists.

## Engineering Handoff
- Implemented 2026-05-03.
- Added `mcpops.Local.IndexRefresh` using the configured `remoteindex.Client`.
- Added `accepted`, `job_id`, `last_refresh_at`, and `message` fields to `IndexRefreshData`.
- Preserved local-only graceful degradation when no remote indexer is configured.
- Added retryable degradation for remote refresh failures.
- Added CLI surface for `cairn index refresh`.
- Added a `sync pull` next step suggesting remote index refresh when a remote indexer is configured.

## QA Handoff
- Accepted 2026-05-03.
- Verified accepted asynchronous refresh reports accepted state and job id.
- Verified completed refresh reports refreshed state and suggests search.
- Verified missing remote indexer degrades with warnings, unavailable mode, and next step.
- Verified failed remote refresh degrades as retryable without throwing.
- Verified common envelope shape and changed paths are preserved.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260503-mcp-mutating-tools-gated`.
