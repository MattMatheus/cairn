# Story: Document Frontmatter And Validation Core

## Metadata
- `id`: STORY-20260502-document-frontmatter-validation
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-document-model-lifecycle]
- `success_metric`: Cairn can parse managed markdown frontmatter, validate core fields, and report warning versus blocking results according to the accepted document lifecycle ADR.
- `release_scope`: required

## Problem Statement
- Cairn needs a shared document model and validation layer before capture, promotion, sync, indexing, and MCP operations can behave consistently.

## Scope
- In:
  - Define core document metadata structs/types.
  - Parse markdown frontmatter.
  - Validate required Cairn core fields.
  - Classify validation findings as error, warning, or info.
  - Distinguish permissive discovery warnings from durable-boundary blocking errors.
  - Add focused tests for valid, missing, invalid, and unknown frontmatter fields.
- Out:
  - Capture or promotion commands.
  - Sync, indexing, or MCP tool implementation.
  - Custom schema validation beyond preserving required Cairn core fields.

## Assumptions
- `docs/adr/ADR-document-model-lifecycle.md` is the source of truth for validation behavior.
- The repository implementation language and package layout may be established or refined during this story.
- Because the repo currently has no product implementation code, this story should create the smallest viable core package/module boundary needed for document metadata and validation.

## Acceptance Criteria
1. Core frontmatter fields from the ADR can be parsed into typed metadata.
2. Missing or invalid required fields produce validation findings with deterministic severity.
3. Unknown frontmatter fields produce warnings, not failures.
4. Validation can be called with a durable-boundary mode that blocks canonical promotion, sync, and indexing.
5. Tests cover representative valid, warning, and blocking cases.

## Validation
- Required checks:
  - Relevant unit tests for frontmatter parsing and validation.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual comparison against `docs/adr/ADR-document-model-lifecycle.md`.

## Dependencies
- Accepted `ADR-document-model-lifecycle`.

## Risks
- Over-strict validation could make ordinary markdown adoption brittle.
- Under-specified parsing behavior could create incompatibility across agents and editors.

## Open Questions
- Resolved for active execution: engineering should choose the smallest idiomatic core package/module for document metadata and validation, then record the chosen path in the QA handoff.

## Next Step
- Engineering should implement the document metadata/frontmatter validation foundation and move the story to engineering QA.

## PM Handoff
- `What changed`: Promoted this story from engineering intake to engineering active as the first implementation slice.
- `Why it matters`: It establishes the shared validation foundation required by lifecycle operations, sync, indexing, and MCP schemas.
- `Acceptance criteria`: Existing criteria remain valid and testable.
- `Risks and assumptions`: Product implementation layout is not established yet; engineering should choose the smallest viable core package/module and document it in the handoff.
- `Completed work summary`: Refined and activated the frontmatter validation core story.
- `Next suggested or required step`: Engineering should implement this active story before dependent lifecycle, sync, or indexing work.
- `Next state recommendation`: engineering active

## Engineering Handoff
- `Changed implementation`: Added the initial Go module and `internal/document` package for markdown frontmatter parsing, typed Cairn metadata, validation modes, severities, and validation findings.
- `Package boundary`: `internal/document` owns document metadata parsing and validation for now. This is the smallest viable core boundary and avoids introducing CLI, MCP, sync, or indexing packages prematurely.
- `Validation coverage`: Added unit tests for valid core frontmatter, missing frontmatter, missing required fields, unknown fields, invalid known values, invalid field types, discovery warnings, and durable-boundary blocking errors.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed.
- `Open risks`: The parser intentionally supports the subset of YAML frontmatter Cairn needs now. If custom schemas require richer YAML features, a later story should evaluate a YAML library or expand parsing behind this package boundary.
- `Assumptions carried forward`: Durable-boundary mode is the shared gate for canonical promotion, sync, and indexing eligibility.
- `QA focus areas`: Confirm validation severity behavior matches `docs/adr/ADR-document-model-lifecycle.md`; confirm unknown fields warn but do not block; confirm missing/invalid fields block only in durable-boundary mode.
- `Action and approval notes`: Highest action class was local write. No risky or sensitive actions; no approval required.
- `Completed work summary`: Implemented document frontmatter parsing and validation core with tests.
- `Next suggested or required step`: QA should review this story, run the Go tests with sandbox-safe `GOCACHE`, and move it to done or return it to active with findings.
- `Next state recommendation`: engineering QA

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed the story acceptance criteria against `internal/document` implementation and tests. The package parses core frontmatter into typed metadata, validates required fields, classifies findings by severity, treats unknown fields as warnings, and distinguishes discovery warnings from durable-boundary blocking errors.
- `Evidence quality call`: Strong enough for this slice. Unit tests cover valid metadata, missing frontmatter, missing required fields, invalid known values, invalid field types, unknown fields, discovery warnings, and durable-boundary errors.
- `Defects`: None.
- `Required fixes`: None.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed. `bash Flywheel/flywheel/tools/validate_intake_items.sh` passed. `git diff --check` passed.
- `Regression risk`: Low. This is the first product code package and has focused tests around the new behavior.
- `Completed work summary`: Accepted the document frontmatter and validation core implementation.
- `Next suggested or required step`: PM should refine the next dependent story, likely capture/promotion/archive lifecycle operations.
- `Next state recommendation`: engineering done
