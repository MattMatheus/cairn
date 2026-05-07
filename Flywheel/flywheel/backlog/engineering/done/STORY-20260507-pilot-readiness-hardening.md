# Story: Pilot Readiness Hardening

## Metadata
- `id`: STORY-20260507-pilot-readiness-hardening
- `owner_role`: Product Manager
- `status`: done
- `source`: human request
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: A maintainer can run one preflight command and a pilot engineer has a clear first-run script, example workspace, and feedback template.
- `release_scope`: Cairn Core v1 pilot

## Problem Statement
- The first engineer pilot is high stakes. Engineers often reject tools that fail or feel confusing on first contact, so Cairn needs a polished, narrow, no-service pilot path with clear expectations and repeatable preflight validation.

## Scope
- In:
  - Add pilot guide and feedback template.
  - Add a small example workspace fixture.
  - Add a one-command pilot readiness check.
  - Keep generated local build/workspace state out of the repo.
  - Remove the local prototype indexer command/service from the active tree so pilots do not infer that a remote indexer is part of v1.
  - Make fresh workspace validation avoid warning about normal first-run missing index/sync state.
- Out:
  - Production packaging.
  - Hosted Azure setup.
  - CocoIndex or semantic retrieval.

## Acceptance Criteria
1. `make pilot-check` passes.
2. Pilot docs explain setup, expected output, sync expectations, conflict demonstration, and known limits.
3. Example workspace validates, indexes, and searches from a copied throwaway location.
4. Root `.gitignore` prevents accidental root `bin/`, `.cache/`, or `.cairn/` debris from appearing as untracked pilot artifacts.
5. The active command/test surface no longer includes `cmd/cairn-indexer`.

## Engineering Handoff
- `Implementation summary`: Added `docs/user/pilot.md`, `docs/user/pilot-feedback.md`, `examples/pilot-workspace/`, `Makefile`, and `deployments/local-dev/pilot-check.sh`. Updated README/user/local-dev docs to point pilots at the new path. Adjusted workspace validation so missing local index and missing sync state are normal before first `index refresh` or first sync. Removed `cmd/cairn-indexer` and the local prototype remote-index service.
- `Action and approval notes`: Local writes and local deletions only. The human explicitly requested doing all pilot-readiness items.
- `Validation`: `make pilot-check` passed.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: `make pilot-check` ran the Go test suite, built a throwaway binary, checked help output, validated/searched the example workspace, and ran the no-service sync/conflict smoke.
- `Residual risks`: The first real pilot may still reveal wording or workflow friction; use `docs/user/pilot-feedback.md` immediately after the session.

## Next Step
- Perform final code review, then commit the v1 pilot-ready changes.
