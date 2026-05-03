# Story: MCP Read Document Progressive Disclosure

## Metadata
- `id`: STORY-20260503-mcp-read-document-progressive
- `owner_role`: Software Architect
- `status`: active
- `source`: pm
- `decision_refs`: [ADR-mcp-operation-surface, ADR-document-model-lifecycle]
- `success_metric`: `read_document` has a local transport-neutral implementation for summary, frontmatter, table-of-contents, sections, and full modes.
- `release_scope`: required

## Problem Statement
- The local MCP read/search adapter intentionally left `read_document` out because progressive markdown parsing is a distinct behavior slice. Agents need compact, mode-based document reads before a server transport is useful.

## Scope
- In:
  - Implement local `read_document` operation in `internal/mcpops`.
  - Resolve documents by id, slug, or path using local metadata where possible.
  - Parse managed markdown frontmatter and body.
  - Support `summary`, `frontmatter`, `toc`, `sections`, and `full` modes.
  - Return `mcpschema.Envelope[mcpschema.ReadDocumentData]` with provenance, warnings, and next steps where helpful.
  - Add tests for each read mode and section-missing behavior.
- Out:
  - MCP server transport.
  - Remote document reads.
  - Rich generated summaries from CocoIndex.
  - Mutating document operations.

## Assumptions
- Local metadata index and local MCP read/search operations are complete.
- Summary mode can use a deterministic local excerpt for v1.
- Section matching can be heading-text based and deterministic for v1.

## Acceptance Criteria
1. `read_document` resolves documents by path, slug, or id.
2. `summary`, `frontmatter`, `toc`, `sections`, and `full` modes return the expected `ReadDocumentData` fields.
3. Section reads return requested sections by heading and warn when a requested heading is absent.
4. Full reads are explicit and include provenance/next-step guidance.
5. Tests cover all read modes and no server transport is introduced.

## Validation
- Required checks:
  - Unit tests for progressive read behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/adr/ADR-mcp-operation-surface.md`.

## Dependencies
- `STORY-20260503-mcp-local-read-search-operations`

## Risks
- Markdown parsing can grow quickly; keep v1 heading parsing simple and deterministic.
- Summary mode should avoid pretending to be semantic summarization.

## Open Questions
- None for this slice.

## Next Step
- Engineering should implement progressive local `read_document` and tests, then move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted progressive `read_document` as the next active engineering slice.
- `Why it matters`: It completes the local read/search MCP operation subset before mutation or transport work.
- `Acceptance criteria`: Focused on mode-based document reads, section lookup, warnings, and transport-neutral implementation.
- `Risks and assumptions`: Keep markdown parsing lightweight; generated/rich summaries are out of scope.
- `Completed work summary`: Added active `read_document` story and created the broader backlog behind it.
- `Next suggested or required step`: Engineering should implement local progressive document reads.
- `Next state recommendation`: engineering active
