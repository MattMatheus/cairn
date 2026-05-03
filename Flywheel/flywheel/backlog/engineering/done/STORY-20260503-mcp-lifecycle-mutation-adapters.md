# Story: MCP Lifecycle Mutation Adapters

## Metadata
- `id`: STORY-20260503-mcp-lifecycle-mutation-adapters
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface, ADR-document-model-lifecycle]
- `success_metric`: Capture, promote, and archive have transport-neutral MCP operation adapters that reuse document lifecycle primitives and response envelopes.
- `release_scope`: required

## Problem Statement
- Document lifecycle primitives exist, but MCP-facing mutation functions do not yet bind schema requests to those operations with consistent response envelopes.

## Scope
- In:
  - Implement local `capture_note`, `promote_document`, and `archive_document` adapters in `internal/mcpops`.
  - Reuse `internal/document.Workspace` lifecycle operations.
  - Return changed paths, durable document ids, warnings, provenance, and next steps.
  - Preserve MCP hard-delete/purge exclusion.
  - Add tests for successful mutations and validation failures.
- Out:
  - MCP server transport.
  - CLI command wiring.
  - Hard delete/purge.
  - Remote sync side effects.

## Assumptions
- Lifecycle operations are the source of truth for document mutation behavior.
- Adapter-level errors should preserve enough detail for MCP response mapping.

## Acceptance Criteria
1. Capture adapter writes under `/agents/{actor}` and returns changed path/id.
2. Promote adapter returns changed path/id and preserves lifecycle validation behavior.
3. Archive adapter returns archive path/id and does not expose purge.
4. All mutation responses use the common envelope and include next steps.
5. Tests cover success and representative failure cases.

## Validation
- Required checks:
  - Unit/integration tests for adapter behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against document lifecycle and MCP ADRs.

## Dependencies
- `STORY-20260502-capture-promotion-archive`
- `STORY-20260502-mcp-schema-surface`

## Risks
- Avoid duplicating lifecycle rules in the adapter.

## Open Questions
- None.

## Next Step
- Engineering should implement local lifecycle mutation adapters and tests, then move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted MCP lifecycle mutation adapters after progressive `read_document` passed QA.
- `Why it matters`: This completes the local MCP operation adapter layer for core document lifecycle mutations while reusing existing document primitives.
- `Acceptance criteria`: Existing criteria remain valid and focused on capture, promote, archive, envelope responses, next steps, and purge exclusion.
- `Risks and assumptions`: Do not duplicate lifecycle rules. Keep MCP server transport, CLI wiring, sync, and hard delete/purge out of scope.
- `Completed work summary`: Activated the lifecycle adapter story.
- `Next suggested or required step`: Engineering should implement local capture/promote/archive adapters in `internal/mcpops`.
- `Next state recommendation`: engineering active

## Engineering Handoff
- `What changed`: Added transport-neutral `CaptureNote`, `PromoteDocument`, and `ArchiveDocument` adapters in `internal/mcpops`, backed by `internal/document.Workspace`.
- `Why it matters`: MCP-facing lifecycle mutations now reuse the document lifecycle source of truth while returning common envelopes with changed paths, durable ids, provenance, and next steps.
- `Acceptance criteria`: Covered capture under `/agents/{actor}`, promote changed path/id behavior, archive path/id behavior, next-step responses, validation error propagation, and no purge/delete adapter.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`; `bash Flywheel/flywheel/tools/validate_intake_items.sh`; `git diff --check`.
- `Risks and assumptions`: Adapter errors currently return as Go errors for future transport mapping. Lifecycle rules remain owned by `internal/document`.
- `QA focus areas`: Verify the adapter does not duplicate lifecycle behavior, does not expose purge/delete, and returns changed paths/ids/next steps for all mutations.
- `Completed work summary`: Implemented MCP lifecycle mutation adapters and tests.
- `Next suggested or required step`: QA should review against lifecycle and MCP ADRs, then move the story to done or file focused gaps.
- `Next state recommendation`: engineering qa

## QA Handoff
- `Verdict`: Pass.
- `Evidence summary`: Capture, promote, and archive adapters delegate to `internal/document.Workspace`, return common envelopes with changed paths, durable ids, provenance, and next steps, propagate lifecycle validation errors, and do not expose purge/delete.
- `Evidence quality call`: Strong for this adapter slice. Tests cover the full capture-promote-archive path and representative validation failures.
- `Defects`: None filed.
- `Required fixes`: None.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`; `bash Flywheel/flywheel/tools/validate_intake_items.sh`; `git diff --check`.
- `Completed work summary`: QA accepted MCP lifecycle mutation adapters.
- `Next suggested or required step`: Close the cycle with an observer report and commit, then promote the next backlog story.
- `Next state recommendation`: engineering done
