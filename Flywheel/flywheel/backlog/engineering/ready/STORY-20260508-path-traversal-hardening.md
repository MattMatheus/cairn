# Story: Path-traversal hardening for repo attach and pull

## Metadata
- `id`: STORY-20260508-path-traversal-hardening
- `owner_role`: SDET
- `status`: ready
- `source`: planning
- `decision_refs`: []
- `success_metric`: No code path can write workspace-controlled files outside the workspace tree on POSIX or Windows.
- `release_scope`: required

## Problem Statement
- `internal/workspace/repos.go:232-238` (`cleanRepoPath`) accepts `..`, `../sibling`, and absolute paths; `AttachRepo` then writes `.cairn-workspace` into the resolved arbitrary directory (`repos.go:104`) (High, H4).
- `internal/syncstate/pull.go:228-234` checks `clean[:3] == "../"` only — Windows escapes via `..\foo` slip through (High, H7).

## Scope
- In:
  - Reject `..`, `../*`, and absolute results in `cleanRepoPath` (mirror `cleanWorkspacePath` in `internal/document/lifecycle.go:484-493`).
  - Replace POSIX-only string check in `pull.go` with `filepath.Rel`-based containment, matching `internal/remotestore/local_fs.go:130-147`.
  - Tests on both forward- and backslash inputs.
- Out:
  - General URL/path policy refactor.

## Assumptions
- `filepath.Rel` is reliable for containment checks on both platforms when both paths are absolute.

## Acceptance Criteria
1. `cleanRepoPath("..")`, `cleanRepoPath("../x")`, `cleanRepoPath("/etc")` all return errors.
2. `pull.go` containment test rejects `..\..\Windows\System32\foo` and `../../etc/passwd` on both POSIX and Windows path forms (use a helper that doesn't depend on `runtime.GOOS`).
3. Existing valid relative paths still pass.

## Validation
- Required checks:
  - `go test ./internal/workspace/...`
  - `go test ./internal/syncstate/...`
  - `make pilot-check`
- Additional checks:
  - Targeted unit tests using mixed separators.

## Dependencies
- None.

## Risks
- Tightening rules may reject legitimate user input; ensure error messages explain rejection.

## Open Questions
- None.

## Next Step
- PM ranks; engineering implements.
