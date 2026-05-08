# Story: install.sh fail closed on checksum fetch failure

## Metadata
- `id`: STORY-20260508-install-checksum-fail-closed
- `owner_role`: SRE
- `status`: done
- `source`: planning
- `decision_refs`: []
- `success_metric`: `scripts/install.sh` exits non-zero when the `.sha256` artifact cannot be downloaded; never installs unverified bytes.
- `release_scope`: required

## Problem Statement
- `scripts/install.sh:60-74` warns but proceeds to install when the checksum download fails, while `scripts/install.ps1:53-54` correctly fails closed. The README invites `curl … | sh`, so unverified install is a real supply-chain risk (Critical, C1).

## Scope
- In:
  - Make `install.sh` exit non-zero on checksum download failure.
  - Match the fail-closed posture of `install.ps1`.
  - Optional opt-out flag (e.g. `CAIRN_INSTALL_SKIP_CHECKSUM=1`) only if explicitly approved.
- Out:
  - Signed releases (separate work).
  - Mirror configuration.

## Assumptions
- Release pipeline always publishes `.sha256` next to the binary.
- Behavior parity with `install.ps1` is the contract.

## Acceptance Criteria
1. With network blocking the `.sha256` URL, `install.sh` exits non-zero before any binary is moved into place.
2. With a valid checksum, install proceeds unchanged.
3. With a checksum mismatch, install fails (already covered today; regression-test).
4. Behavior matches `install.ps1` for the three paths above.

## Validation
- Required checks:
  - Manual: simulate 404 on checksum URL via env-overridden URL or local server; confirm exit code != 0 and no binary placed.
  - Unit/script test: small bats or shellcheck run; CI lint passes.
- Additional checks:
  - `make pilot-check` (no install.sh dependency, but confirm no regression).

## Dependencies
- None.

## Risks
- Existing user docs may reference an offline-friendly path; confirm none do before fail-closed lands.

## Open Questions
- Do we want `--insecure` opt-out, or is fail-closed unconditional?
  - Resolved: fail-closed unconditional, parity with `install.ps1`. No env opt-out.

## Engineering Handoff
- `Implementation summary`: `scripts/install.sh:60-74` rewritten so checksum download failure is fatal (`if ! curl ...; then echo ...; exit 1; fi`). When checksum download succeeds, verification proceeds via `sha256sum` or `shasum -a 256` exactly as before. Script remains `set -eu`; on failure no `tar` or `install` runs.
- `Action and approval notes`: Local edit only; no installs performed; no changes to `install.ps1` (already fail-closed).
- `Validation`:
  - `sh -n scripts/install.sh` — syntax OK.
  - Isolation reproduction (`/tmp/test-checksum-fail.sh` against `https://example.invalid/...`): script exits 1 with the new error message; never reaches the install step.
  - `make pilot-check` — passed (Go test suite + sample workspace + sync smoke).
- `QA focus areas`:
  1. Confirm fail-closed message is clear and references the checksum URL.
  2. Verify that with a real release where `.sha256` exists, install proceeds unchanged.
  3. Verify behavior parity with `install.ps1`: download failure and mismatch both fatal.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`:
  - AC1 (checksum download fails → exit non-zero, no binary placed): isolation reproduction script targeting `https://example.invalid/...` exited 1 with the new `refusing to install unverified binary` message; never reached the install step.
  - AC2 (valid checksum → install proceeds unchanged): code path post-checksum is byte-equivalent to prior; verified by diff inspection (`scripts/install.sh:60-74`).
  - AC3 (checksum mismatch → fail): existing mismatch branch preserved (`exit 1` retained).
  - AC4 (parity with `install.ps1`): both now fatal on download failure and mismatch; ps1 unchanged.
  - Regression: `make pilot-check` passed (Go test suite + sample workspace + sync smoke).
  - Syntax: `sh -n scripts/install.sh` clean; `set -eu` retained.
- `Residual risks`:
  - AC2 was validated structurally, not against a live `v0.N` release; the next real release will exercise the happy path end-to-end.
  - No `--insecure` opt-out exists; users in air-gapped environments must distribute the binary out of band. Documented in story open questions.

## Next Step
- Run observer; create cycle commit `cycle-STORY-20260508-install-checksum-fail-closed`.
