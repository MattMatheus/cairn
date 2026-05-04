# Story: MCP Read Document Progressive Disclosure

## Metadata
- `id`: STORY-20260503-mcp-read-document-progressive
- `owner_role`: Software Architect
- `status`: done
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

## Engineering Handoff
- `What changed`: Added local `ReadDocument` support in `internal/mcpops` with deterministic markdown body parsing, heading TOC extraction, section selection, summary excerpt, frontmatter response, and explicit full-content mode.
- `Why it matters`: Agents can now progressively read managed markdown through the MCP operation layer without requiring server transport or rich semantic summaries.
- `Acceptance criteria`: Covered path/slug/id resolution; `summary`, `frontmatter`, `toc`, `sections`, and `full` modes; missing-section warnings; provenance and next-step guidance for progressive/full reads.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`; `bash Flywheel/flywheel/tools/validate_intake_items.sh`; `git diff --check`.
- `Risks and assumptions`: Heading parsing is intentionally lightweight and deterministic. Summary mode is an excerpt, not a generated summary.
- `QA focus areas`: Verify all read modes, missing-section warnings, id/slug/path resolution, and that no MCP transport or remote reads were introduced.
- `Completed work summary`: Implemented progressive local `read_document` and tests.
- `Next suggested or required step`: QA should review against `docs/adr/ADR-mcp-operation-surface.md`, then move the story to done or file focused bugs.
- `Next state recommendation`: engineering qa

## QA Handoff
- `Verdict`: Pass.
- `Evidence summary`: `ReadDocument` supports path, slug, and id resolution plus `summary`, `frontmatter`, `toc`, `sections`, and `full` modes. Missing requested sections produce warnings, full reads include next-step guidance, and the implementation remains local and transport-neutral.
- `Evidence quality call`: Strong for this lightweight progressive-read slice. Tests cover all read modes and missing-section behavior.
- `Defects`: None filed.
- `Required fixes`: None.
- `Validation`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...`; `bash Flywheel/flywheel/tools/validate_intake_items.sh`; `git diff --check`.
- `Completed work summary`: QA accepted progressive local `read_document`.
- `Next suggested or required step`: Close the cycle with an observer report and commit, then PM can promote the next backlog story.
- `Next state recommendation`: engineering done
