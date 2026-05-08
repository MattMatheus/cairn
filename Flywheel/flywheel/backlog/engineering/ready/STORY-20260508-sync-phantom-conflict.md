# Story: Eliminate phantom conflict on diverged-but-non-overlapping syncs

## Metadata
- `id`: STORY-20260508-sync-phantom-conflict
- `owner_role`: Software Architect
- `status`: ready
- `source`: planning
- `decision_refs`: [ARCH-20260502-sync-conflict-behavior]
- `success_metric`: `compare()` never reports a `Conflict` whose Local and Remote refer to non-overlapping paths/documents.
- `release_scope`: required

## Problem Statement
- `internal/syncstate/compare.go:168-170` synthesizes a fake `Conflict{Local: localChanges[0], Remote: remoteChanges[0]}` when both sides have changes but no overlap. This pairs unrelated edits and `PlanFromStatus` (`plan.go:26-32`) then refuses sync (High, H1).

## Scope
- In:
  - Remove the synthesized conflict.
  - Surface "diverged with no overlap" as a distinct state for the planner and CLI output.
  - Update `PlanFromStatus` so non-overlapping diverged state plans an apply, not a block.
  - Update CLI output so users see "remote and local both changed; no overlap, applying both" rather than a misleading conflict report.
- Out:
  - Three-way merge.
  - New conflict UX beyond clearer messaging.

## Assumptions
- Existing sync semantics permit applying both sides when paths/documents are disjoint.

## Acceptance Criteria
1. Unit test: local and remote both have one independent change with disjoint paths → `Conflicts` is empty; `Diverged` is true; planner returns an apply plan (not a refusal).
2. Existing overlap-conflict tests still produce a `Conflict` entry.
3. CLI status message for diverged-no-overlap is updated and asserted via test or golden output.

## Validation
- Required checks:
  - `go test ./internal/syncstate/...`
  - `go test ./internal/cli/...`
  - `make pilot-check`
- Additional checks:
  - Manual sync smoke against local-fs remote store with disjoint changes.

## Dependencies
- None.

## Risks
- Plan change may surprise users relying on prior "blocks on diverge" behavior; mitigate with clear status output.

## Open Questions
- Should diverged-no-overlap require explicit `--apply-diverged` flag for safety?

## Next Step
- PM ranks; engineering implements.
