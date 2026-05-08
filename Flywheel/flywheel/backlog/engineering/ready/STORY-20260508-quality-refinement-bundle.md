# Story: Quality refinement bundle (Medium + selected Low findings)

## Metadata
- `id`: STORY-20260508-quality-refinement-bundle
- `owner_role`: Software Architect
- `status`: ready
- `source`: planning
- `decision_refs`: []
- `success_metric`: All bundled fixes land with tests; no regression in `make pilot-check`.
- `release_scope`: deferred

## Problem Statement
- Eleven Medium and several Low findings from `findings-5-8.md` are individually small but collectively shape code quality, performance, and small-input correctness. Bundling avoids cycle thrash on tiny stories.

## Scope
- In (each bullet maps to a finding):
  - M1: gate hardcoded Azurite key behind explicit env or comment.
  - M2: rename shadowed `results` in `internal/localindex/search.go:43-57`.
  - M3: UTF-8-safe slicing in `snippet`/`summarize`.
  - M4: escape `%`/`_`/`\` in LIKE-bound metadata queries.
  - M5: replace `strings.Trim` cutset with delimiter strip in `unquoteConfig` (`config.go:614-619`) and `repos.go:251-256`.
  - M6: hoist `LoadConfig` out of per-file walk in `validate.go` and `health.go`.
  - M7: normalize bad-but-non-zero values in `repairMetadata`.
  - M8: collapse redundant status assignment in `lifecycle.go:142-161`.
  - M9: add inter-phase ctx checks in `health.go`.
  - M10: `O_EXCL` on ADR creation in `nextADRNumber`.
  - M11: skip auto-reindex in `runSearch` when index already populated.
  - L1: quoted-comma support in `splitCSV`.
  - L2: typed sentinel for "remote store is required" sync error.
  - L8: drop unused `path` import in `extensions/vscode-cairn/src/extension.js`.
- Out:
  - L3, L4, L5, L6, L7 (style/cosmetic; defer to a future cleanup pass).
  - Anything not listed above.

## Assumptions
- All listed items can be implemented and reviewed within one engineering cycle by a senior engineer.

## Acceptance Criteria
1. Each in-scope item has a code change and at least one focused test or test update.
2. `go test ./...` passes.
3. `make pilot-check` passes.
4. PR description maps each commit/section back to its finding ID.

## Validation
- Required checks:
  - `go test ./...`
  - `make pilot-check`
- Additional checks:
  - `go vet ./...`
  - Spot-run of `cairn search` and `cairn note` against the example workspace.

## Dependencies
- Should land after the Critical/High stories to avoid merge churn in the same files.

## Risks
- Medium items collide on `lifecycle.go` and `cli.go` with H6 and H1 stories; sequence carefully.

## Open Questions
- Split bundle if it exceeds a single cycle? Decide at engineering kickoff.

## Next Step
- PM ranks last in priority; engineering picks up after Critical/High stories close.
