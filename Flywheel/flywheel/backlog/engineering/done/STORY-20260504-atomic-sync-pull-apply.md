# Story: Atomic Sync Pull Apply

## Metadata
- `id`: STORY-20260504-atomic-sync-pull-apply
- `owner_role`: SRE
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-sync-conflict-behavior]
- `success_metric`: Pull failures do not leave accepted sync state behind and either avoid or clearly recover partial local file mutations.
- `release_scope`: required

## Problem Statement
- `ApplyPull` stages remote object content, then writes/removes workspace paths one by one before saving sync state. If a later file operation fails, earlier files remain changed while the sync state still points at the previous base.

## Scope
- In:
  - Make pull apply resilient to mid-apply failures.
  - Prefer staging writes and atomic rename where practical.
  - Avoid destructive removes until replacement writes are staged.
  - Return clear recovery warnings if rollback is impossible.
  - Add tests for write failure and remove failure paths.
- Out:
  - Automatic merge.
  - Transactional filesystem abstraction across all platforms beyond the story scope.

## Assumptions
- Windows support matters, so atomic file operations should use portable Go behavior or platform-conscious fallbacks.

## Acceptance Criteria
1. Pull does not save new sync state if any file mutation fails.
2. Pull avoids partially written file contents by writing staged content before replacing targets.
3. Failure tests prove prior files are not corrupted when a later operation fails, or that recovery guidance is explicit when rollback is not possible.
4. Move/archive/delete pull cases remain covered.
5. Existing safe pull behavior remains unchanged.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/syncstate ./internal/mcpops`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Add fake filesystem or permission-based failure tests for staged pull apply.

## Dependencies
- STORY-20260504-sync-live-remote-manifest-safety

## Risks
- Full rollback may be hard to guarantee on all filesystems; the implementation must be honest about residual risk.

## Open Questions
- Is best-effort rollback sufficient for v1, or should pull require a pre-apply workspace backup manifest?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: Pull apply now backs up affected paths, writes pulled content through temp files, rolls back affected paths on apply failure, and avoids saving new sync state when apply fails.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/syncstate ./internal/mcpops`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm injected mid-apply write failure restores earlier file content and does not advance sync state.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Added rollback regression coverage for an injected later write failure. Existing move/archive/delete pull tests and full suite passed.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
