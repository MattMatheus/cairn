# Story: Knowledge Health Report

## Metadata
- `id`: STORY-20260508-knowledge-health-report
- `owner_role`: Product Manager
- `status`: intake
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

Keep deferred until pilot readiness and capture polish are complete.
