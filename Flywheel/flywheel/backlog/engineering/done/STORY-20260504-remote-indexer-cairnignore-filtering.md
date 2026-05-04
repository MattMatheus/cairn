# Story: Remote Indexer Cairnignore Filtering

## Metadata
- `id`: STORY-20260504-remote-indexer-cairnignore-filtering
- `owner_role`: SRE
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-indexing-query-boundary]
- `success_metric`: Remote indexer refresh and search exclude `.cairnignore` paths and match local workspace filtering behavior.
- `release_scope`: required

## Problem Statement
- The prototype remote indexer walks markdown under the workspace while only skipping `.git` and `.cairn`. It does not honor `.cairnignore`, so ignored markdown can enter remote search results.

## Scope
- In:
  - Reuse or share the workspace ignore matcher for the remote indexer.
  - Skip ignored directories and files during refresh/count/search.
  - Keep `.git` and `.cairn` exclusions.
  - Add tests proving ignored markdown is not counted or returned.
- Out:
  - Semantic embedding implementation.
  - Remote authorization changes.

## Assumptions
- Remote indexer and local indexer should agree on document eligibility for ignored paths.

## Acceptance Criteria
1. Remote indexer excludes files and directories matched by `.cairnignore`.
2. Ignored managed markdown does not affect indexed count.
3. Ignored managed markdown is not returned from `/search`.
4. Tests cover ignored directory and ignored file patterns.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/remoteindex ./internal/localindex`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Compare local and remote prototype search results for a fixture workspace with ignored markdown.

## Dependencies
- None.

## Risks
- Duplicated ignore implementations can drift; prefer shared code or matching test fixtures.

## Open Questions
- Should remote indexing also enforce configured managed-folder membership, or only `.cairnignore` plus valid frontmatter?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: Remote indexer search/count now load `.cairnignore` and skip ignored directories and files, while preserving `.git` and `.cairn` exclusions.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/remoteindex ./internal/localindex`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm ignored markdown is neither counted by refresh nor returned by search.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Added remote indexer test for ignored directory and file patterns. Targeted and full suites passed.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
