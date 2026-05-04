# Story: Prune Local Index Stale Rows

## Metadata
- `id`: STORY-20260504-prune-local-index-stale-rows
- `owner_role`: Database Expert
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Local index refresh removes rows for deleted, ignored, or no-longer-managed documents.
- `release_scope`: required

## Problem Statement
- `IndexWorkspace` upserts discovered managed markdown but never removes stale database rows. Search, list, and MCP document resolution can return paths for files that no longer exist or should no longer be indexed.

## Scope
- In:
  - Track the set of paths discovered during `IndexWorkspace`.
  - Delete rows for paths not discovered or no longer valid after a full workspace refresh.
  - Preserve single-file indexing behavior if it remains useful.
  - Add tests for deleted files, newly ignored files, and documents that lose valid managed frontmatter.
- Out:
  - Incremental indexing scheduler.
  - FTS virtual table redesign.

## Assumptions
- `IndexWorkspace` is a full refresh and may prune stale rows safely.

## Acceptance Criteria
1. Deleted markdown documents disappear from query results after `IndexWorkspace`.
2. Newly ignored markdown documents disappear from query results after `IndexWorkspace`.
3. Markdown that loses valid managed frontmatter disappears from query results after `IndexWorkspace`.
4. Existing metadata and full-text search tests continue to pass.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/localindex ./internal/mcpops`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Verify `read_document` by stale path fails after refresh rather than returning old index metadata.

## Dependencies
- None.

## Risks
- Care is needed not to prune rows during targeted single-file indexing.

## Open Questions
- Should skipped invalid managed files be reported separately from pruned rows in `IndexReport`?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: Full local index refresh prunes database rows not rediscovered as valid indexed documents. Added coverage for deleted, newly ignored, and invalidated markdown paths.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/localindex ./internal/mcpops`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm `IndexMarkdownFile` remains targeted while `IndexWorkspace` is the full-refresh prune boundary.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Added stale-row pruning tests and full suite passed.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
