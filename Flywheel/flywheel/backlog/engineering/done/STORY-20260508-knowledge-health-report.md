# Story: Knowledge Health Report

## Metadata
- `id`: STORY-20260508-knowledge-health-report
- `owner_role`: Product Manager
- `status`: done
- `source`: planning
- `decision_refs`: []
- `success_metric`: A pod can generate a markdown or CLI health report showing actionable Cairn knowledge base counts and stale states without a dashboard.
- `release_scope`: deferred

## Problem Statement

Pods and the central AI team will eventually need light visibility into Cairn health, but a dashboard and telemetry platform are premature. A local report can expose useful counts and stale states while keeping the product small.

## Scope
- In: Generate a CLI and/or markdown health report from local workspace state.
- In: Include document counts by type/status, proposed docs awaiting review, stale working docs, validation findings, recent canonical promotions, and index/sync freshness.
- Out: Hosted dashboards, cross-pod search, central telemetry, or automatic scoring that drives governance.

## Assumptions

- Health can be computed from local documents, frontmatter, index state, and sync metadata.
- The first version should be useful for pilots and future dashboard design without becoming the dashboard.

## Acceptance Criteria
1. A command can print or generate a health report for a local Cairn workspace.
2. The report identifies actionable stale or invalid states without requiring remote services.
3. The output is stable enough to use in pilot feedback and future product planning.

## Validation
- Required checks: `go test ./...`
- Additional checks: fixture workspace with mixed statuses and validation findings.

## Dependencies

- Existing document listing, validation, index, and sync status operations.

## Risks

- Health reports can imply governance or scoring before teams trust the tool. Keep language descriptive and actionable.
- Counts can become misleading if ignored files or unmanaged markdown are not clearly excluded.

## Open Questions

- Should generated markdown land under `.cairn/generated/`, `onboarding/`, or be stdout-only in v1?

## Next Step

Engineering should implement a local markdown health report command.

## PM Refinement

- Command shape: `cairn health report [--output PATH]`.
- Default behavior should print markdown to stdout.
- `--output` may write the same markdown report to a workspace-relative path.
- The report should be descriptive, not a score or governance mechanism.
- Keep all data local: no dashboard, central telemetry, or remote service dependency.

## Engineering Handoff

### What Changed

- Added `cairn health report`.
- Added `cairn health report --output PATH` for workspace-relative markdown output.
- Added local health report builder and renderer.
- Report includes:
  - managed document total
  - counts by status and type
  - proposed documents awaiting review
  - stale working documents
  - recent canonical documents
  - validation findings
  - local index availability
  - sync divergence/change/conflict counts
  - descriptive follow-up suggestions
- The report is local-only and descriptive; it does not score, govern, publish telemetry, or call remote services.

### Validation Evidence

- `go test ./...` passed.
- Manual mixed-workspace smoke printed a report with proposed document, validation warning, missing index, sync counts, and follow-up suggestions.
- Manual `--output .cairn/generated/health.md` smoke wrote markdown successfully.

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, production, dashboard, telemetry, or remote actions were taken.
- No human approval was required beyond the user's request to finish the final story.

### Risks And Assumptions

- Stale working documents are defined as `working` documents older than 30 days.
- The report counts managed markdown with valid Cairn frontmatter; unmanaged or invalid markdown is represented through validation findings rather than document counts.
- Sync counts use local manifest comparison and do not require a remote service.

### QA Focus Areas

- Verify stdout report includes the required health sections.
- Verify `--output` writes a markdown file.
- Verify a mixed fixture reports proposed documents and validation findings.
- Confirm the language remains descriptive and does not imply governance scoring.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `go test ./...` passed.
- Mixed workspace smoke printed a report with proposed document, validation warning, missing index, sync counts, and follow-up suggestions.
- `--output .cairn/generated/health.md` smoke wrote a non-empty markdown report.

### Defects

None filed.

### Evidence Quality

Sufficient. Automated tests cover report building/rendering and CLI stdout/output paths; manual smokes cover mixed health states.

### State Transition

Moved to `done`.
