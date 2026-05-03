# Bug: Capture Can Create Invalid Frontmatter

## Metadata
- `id`: BUG-20260502-capture-invalid-frontmatter
- `priority`: P1
- `reported_by`: QA Engineer
- `source_story`: STORY-20260502-capture-promotion-archive
- `status`: done
- `decision_refs`: [ADR-document-model-lifecycle]
- `impact_metric`: Capture operation can violate the story requirement that captured documents have valid core frontmatter.

## Priority Definitions
- `P0`: release-blocking, data loss/corruption, or security-critical
- `P1`: major functional regression or blocked acceptance criteria
- `P2`: moderate defect with workaround
- `P3`: minor defect, polish issue, or low-impact inconsistency

## Summary
`Workspace.Capture` writes metadata without validating the generated frontmatter. Invalid `Type` values, titles that slugify to an empty slug, or unsafe path-ish actor values can produce documents that fail durable-boundary validation or land in surprising paths.

## Expected Behavior
- Capture should only create managed markdown with valid core frontmatter.
- Invalid document type, empty generated slug, invalid tags, or unsafe actor/path inputs should return an error before writing a file.

## Actual Behavior
- Capture accepts `CaptureOptions.Type` without checking it against built-in document types.
- Capture uses `slugify(Title)` without checking that the result is non-empty.
- Capture uses `Actor` as a path segment without normalizing or rejecting path traversal-like values.

## Reproduction Steps
1. Call `Workspace.Capture` with `Type: "not-a-type"` and otherwise valid options.
2. Read the created markdown.
3. Validate it with `ValidationModeDurableBoundary`.

## Evidence
- Code review of `internal/document/lifecycle.go` shows capture writes metadata directly without calling `Validate`.
- Existing tests cover a valid capture path but do not cover invalid capture inputs.

## Constraints
- Keep CLI, MCP, sync, and purge out of scope.
- Preserve the reusable operation-function boundary in `internal/document`.

## Risks
- Agents may capture documents that later fail promotion, sync, or indexing eligibility.

## Suggested Fix Direction
- Validate capture-generated metadata with durable-boundary validation before writing.
- Reject invalid document types and empty slugs.
- Treat actor as an identifier/path segment and reject absolute paths, `..`, separators, or values that would move output outside `/agents/{actor}/`.
- Add tests for invalid type, empty slug, invalid tags, and unsafe actor values.

## Next Step
- Fixed by `STORY-20260502-capture-promotion-archive`; no further action required.

## Resolution
- `Verdict`: Fixed.
- `Evidence summary`: `Workspace.Capture` now rejects invalid generated frontmatter before writing. Regression tests cover invalid document type, empty generated slug, invalid tag, unsafe Unix-style actor path, unsafe Windows-style actor path, and no markdown file written on invalid input.
- `Validation results`: `GOCACHE=/private/tmp/cairn-go-cache go test -count=1 ./...` passed.
- `Completed work summary`: Closed capture invalid-frontmatter bug.
- `Next suggested or required step`: Continue with the next PM-refined engineering story.
