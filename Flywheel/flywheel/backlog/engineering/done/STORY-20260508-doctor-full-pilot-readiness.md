# Story: Doctor Full Pilot Readiness

## Metadata
- `id`: STORY-20260508-doctor-full-pilot-readiness
- `owner_role`: SDET
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: A pilot workspace can run one command that reports config, validation, index, sync, remote, and MCP readiness with actionable next steps.
- `release_scope`: required

## Problem Statement

Pilots need one trustworthy command that explains whether a Cairn workspace is ready for real use. Today `doctor`, `validate`, `index`, `sync`, and MCP checks are split across multiple commands, which creates friction and support ambiguity.

## Scope
- In: Add a full doctor or pilot readiness mode that aggregates local workspace, schema, validation, index/search, sync, remote reachability, MCP tool availability, and lifecycle warning checks.
- In: Keep output concise, actionable, and safe for local pilot use.
- Out: Product dashboard, telemetry platform, remote fleet monitoring, or automatic repair of risky conditions.

## Assumptions

- Existing commands and internal operations can be reused rather than duplicating validation logic.
- Remote checks must remain optional or gracefully skipped when no remote sync is configured.
- MCP readiness can be checked by instantiating/listing local tool surfaces without requiring an external MCP client.

## Acceptance Criteria
1. A new full doctor mode reports workspace config/folders, schema validity, validation status, local index status, search sanity, sync status, optional remote reachability, and MCP tool availability.
2. The command prints clear next steps for missing config, stale index, validation failures, missing remote configuration, unreachable remote, or unavailable MCP readiness.
3. The mode is covered by tests or a local smoke path that exercises healthy and degraded workspaces.

## Validation
- Required checks: `go test ./...`
- Additional checks: run the new command against a temporary initialized workspace and one degraded workspace.

## Dependencies

- Existing `doctor`, `validate`, `index`, `sync`, and MCP server construction paths.

## Risks

- The command could become noisy if it prints raw details from every subsystem. Favor summary plus actionable findings.
- Remote reachability can be slow or auth-dependent. Keep remote checks bounded and explicit.

## Open Questions

- Should the command be `cairn doctor --full`, `cairn doctor --pilot`, or both with one alias?

## Next Step

QA should validate the full readiness output against healthy, missing-config, and degraded workspace paths.

## Engineering Handoff

### What Changed

- Added `cairn doctor --full` as an aggregated pilot readiness report.
- Preserved existing `cairn doctor` behavior.
- Kept remote store reachability behind `--remote`; without it, full doctor reports the configured store as skipped.
- Added readiness checks for config, managed folders, schema/config validity, workspace validation, pre-existing local index, search query path, sync status, optional remote reachability, and MCP tool surfaces.
- Added CLI tests for missing config, healthy local-sync workspace, and validation-warning workspace.

### Validation Evidence

- `go test ./...` passed.
- Healthy smoke passed with temporary workspace, `setup local-sync`, capture, `index refresh`, and `doctor --full --remote`.
- Degraded smoke passed with temporary initialized workspace, unmanaged runbook markdown, and `doctor --full`.

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, or production actions were taken.
- No human approval was required beyond the user's request to continue into Engineering and QA.

### Risks And Assumptions

- `doctor --full` checks local index presence before opening the local operation layer, because opening that layer can create generated SQLite state.
- Remote reachability is explicit via `--remote` to avoid accidental Azure/auth/network work during routine local readiness checks.
- Search sanity currently verifies the local query path when an index already exists; it does not assert search quality.

### QA Focus Areas

- Verify missing-config output stays actionable and does not try to open workspace operations.
- Verify healthy local-sync output includes all expected readiness categories.
- Verify degraded workspace output reports validation warnings and missing index next steps.
- Confirm existing `doctor` and `doctor --remote` behavior remains compatible.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `go test ./...` passed.
- Missing-config smoke passed with `doctor --full` reporting skipped dependent checks and `cairn init` next step.
- Healthy local-sync smoke passed with `setup local-sync`, capture, `index refresh`, and `doctor --full --remote`.
- Degraded workspace smoke passed with validation warning, missing index next step, skipped search sanity, and missing remote-sync guidance.

### Defects

None filed.

### Evidence Quality

Sufficient. Coverage includes automated CLI tests and manual QA smokes for healthy, missing-config, and degraded workspaces.

### State Transition

Moved to `done`.
