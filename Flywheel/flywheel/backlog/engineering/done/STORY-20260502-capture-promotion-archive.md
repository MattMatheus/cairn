# Story: Capture Promotion And Archive Lifecycle Operations

## Metadata
- `id`: STORY-20260502-capture-promotion-archive
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-document-model-lifecycle, ADR-mcp-operation-surface]
- `success_metric`: Cairn can create, promote, and archive managed documents while preserving durable identity and lifecycle rules.
- `release_scope`: required

## Problem Statement
- Agents and humans need first-class lifecycle operations instead of ad hoc file writes so canonical knowledge, ADR numbering, and archive behavior stay reliable.

## Scope
- In:
  - Implement capture operation for agent-authored notes under `/agents/{actor}/...`.
  - Generate durable Cairn document ids.
  - Implement promotion to `proposed` and `canonical`.
  - Preserve durable metadata during promotion.
  - Move promoted documents to configured destination folders.
  - Assign ADR numbers when `decision` documents become `canonical`.
  - Implement archive operation preserving original path under `/archive`.
  - Add tests for identity preservation, status transitions, promotion moves, ADR numbering, and archive moves.
- Out:
  - Hard delete or purge implementation.
  - Azure sync behavior.
  - MCP server wiring beyond exposing reusable operation functions.

## Assumptions
- Frontmatter parsing and validation are available from `STORY-20260502-document-frontmatter-validation`.
- ADR number allocation can start with local deterministic behavior and later be guarded by sync conflict handling.
- This story should extend the existing `internal/document` package or add the smallest adjacent package needed for filesystem lifecycle operations.
- MCP and CLI adapters should remain out of scope; expose reusable operation functions that later adapters can call.

## Acceptance Criteria
1. Capture creates a managed markdown document with valid core frontmatter.
2. Promotion to `proposed` repairs or adds required frontmatter when possible.
3. Promotion to `canonical` blocks until core frontmatter is valid.
4. Canonical decision promotion assigns an ADR number and final `/decisions/ADR-000N-slug.md` path.
5. Archive sets `status: archived` and moves the document under `/archive` preserving original path.
6. Tests cover capture, proposed promotion, canonical promotion, ADR numbering, and archive behavior.

## Validation
- Required checks:
  - Relevant unit/integration tests for lifecycle operations.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-document-model-lifecycle.md`.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`

## Risks
- ADR number allocation may need extra locking once sync is implemented.
- Promotion moves can surprise users if paths are not reported clearly.

## Open Questions
- Resolved for this slice: CLI-only purge is deferred until after archive is working and should be created as a separate story when needed.

## Next Step
- Engineering should implement capture, promotion, ADR numbering, and archive lifecycle operations, then move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted this story from engineering intake to engineering active as the next dependency after frontmatter validation.
- `Why it matters`: It turns the accepted document lifecycle ADR into reusable product operations for later CLI and MCP adapters.
- `Acceptance criteria`: Existing criteria remain valid and testable. Purge remains out of scope.
- `Risks and assumptions`: ADR number allocation may be local-only in this slice and later protected by sync conflict behavior. File moves should report paths clearly in operation results.
- `Completed work summary`: Refined and activated the capture, promotion, and archive lifecycle story.
- `Next suggested or required step`: Engineering should implement reusable lifecycle operations on top of `internal/document`.
- `Next state recommendation`: engineering active

## Engineering Handoff
- `Changed implementation`: Added reusable lifecycle operations to `internal/document`: workspace capture, promotion to `proposed` or `canonical`, local ADR number assignment for canonical decisions, archive moves, frontmatter rendering, path slugging, and operation result metadata/next steps.
- `Package boundary`: Lifecycle operations remain in `internal/document` because they directly manipulate document metadata and frontmatter. CLI, MCP, sync, and indexing adapters remain out of scope.
- `Validation coverage`: Added tests for capture, proposed promotion frontmatter repair, durable metadata preservation during proposed promotion, canonical promotion blocking invalid frontmatter, canonical decision ADR numbering/path moves, archive path preservation, and workspace path traversal rejection.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed. `git diff --check` passed.
- `Open risks`: ADR number allocation is local and filesystem-scanning for this slice. Sync-aware locking/conflict handling remains deferred to sync work.
- `Assumptions carried forward`: CLI-only purge remains out of scope. Later adapters should call these reusable operations rather than reimplementing lifecycle file moves.
- `QA focus areas`: Confirm canonical promotion blocks invalid frontmatter; confirm proposed promotion repairs missing frontmatter; confirm ADR numbering uses the next local ADR number; confirm archive moves preserve original path below `/archive`; confirm path traversal is rejected.
- `Action and approval notes`: Highest action class was local write. No risky or sensitive actions; no approval required.
- `Completed work summary`: Implemented capture, promotion, ADR numbering, and archive lifecycle operations with tests.
- `Next suggested or required step`: QA should review this story, run the Go tests with sandbox-safe `GOCACHE`, and move it to done or return it to active with findings.
- `Next state recommendation`: engineering QA

## QA Review
- `Verdict`: Fail.
- `Evidence summary`: Automated checks passed, and most acceptance criteria have direct test coverage. QA review found a blocking gap in acceptance criterion 1: capture can create invalid core frontmatter because `CaptureOptions.Type`, generated slug, tags, and actor path segment are not validated before writing.
- `Evidence quality call`: Strong enough to return to active. The gap is visible by code review and is not covered by current tests.
- `Defects`: `Flywheel/flywheel/backlog/engineering/intake/BUG-20260502-capture-invalid-frontmatter.md`
- `Required fixes`: Capture must reject invalid document type, empty generated slug, invalid tags, and unsafe actor/path inputs before writing. Add tests for these cases.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed. `git diff --check` passed.
- `Regression risk`: Moderate. The bug affects the first write operation agents will use and can create documents that later fail promotion, sync, or indexing eligibility.
- `Completed work summary`: QA reviewed lifecycle operations and identified one blocking capture-validity defect.
- `Next suggested or required step`: Engineering should fix the capture validation defect, rerun tests, and move the story back to QA.
- `Next state recommendation`: engineering active

## Engineering Fix Handoff
- `Changed implementation`: Fixed `Workspace.Capture` to reject invalid generated frontmatter before writing. Capture now validates actor as a single safe path segment, rejects titles that produce empty slugs, validates rendered frontmatter in durable-boundary mode, and reports blocking validation findings.
- `Validation coverage`: Added regression tests for invalid document type, empty slug, invalid tag, and unsafe actor input. Tests assert invalid capture inputs write no markdown files.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed. `git diff --check` passed.
- `Defect addressed`: `Flywheel/flywheel/backlog/engineering/intake/BUG-20260502-capture-invalid-frontmatter.md`
- `Open risks`: Local ADR number allocation still depends on future sync conflict handling, unchanged from the original handoff.
- `QA focus areas`: Recheck capture validity edge cases, especially invalid type, empty slug, invalid tag, and unsafe actor values.
- `Completed work summary`: Fixed the capture validation QA miss and added regression tests.
- `Next suggested or required step`: QA should rerun the lifecycle checks and either accept the story or return it with any remaining findings.
- `Next state recommendation`: engineering QA

## QA Re-Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed the capture validation fix and lifecycle operations against all acceptance criteria. Capture now rejects invalid generated frontmatter before writing; promotion repairs missing frontmatter for proposed documents; canonical promotion blocks invalid frontmatter; decision canonical promotion assigns the next local ADR number and final path; archive moves documents under `/archive` while preserving original path.
- `Evidence quality call`: Strong enough for acceptance. Tests cover the original lifecycle behavior plus the QA regression cases for invalid capture type, empty slug, invalid tags, unsafe Unix-style actor path, unsafe Windows-style actor path, and no-file-written behavior on invalid capture.
- `Defects`: `BUG-20260502-capture-invalid-frontmatter` fixed.
- `Required fixes`: None.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed. `git diff --check` passed.
- `Regression risk`: Low to moderate. Lifecycle operations are well-covered for this slice; ADR number allocation remains local-only until sync conflict handling.
- `Completed work summary`: Accepted capture, promotion, ADR numbering, and archive lifecycle operations.
- `Next suggested or required step`: PM should refine the sync manifest/state story next.
- `Next state recommendation`: engineering done
