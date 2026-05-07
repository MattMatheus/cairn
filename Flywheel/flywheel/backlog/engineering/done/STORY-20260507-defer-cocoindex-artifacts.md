# Story: Defer CocoIndex Artifacts

## Metadata
- `id`: STORY-20260507-defer-cocoindex-artifacts
- `owner_role`: Product Manager
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: Backlog and local-dev artifacts no longer present CocoIndex integration as required v1 work.
- `release_scope`: deferred

## Problem Statement
- CocoIndex work is valuable, but it should not define the v1 critical path. Existing stories, docs, and local-dev artifacts may confuse new contributors by making rich retrieval look mandatory.

## Scope
- In:
  - Reclassify CocoIndex, pgvector, remote indexer, and production remote-index auth work as deferred or optional.
  - Update Flywheel queue notes so the next active batch is Cairn Core v1.
  - Preserve useful reference artifacts for future rich-retrieval work.
  - Ensure no active required v1 story depends on `deployments/cocoindex-indexer`.
- Out:
  - Deleting reference repositories.
  - Removing optional remote index code if it is already implemented and harmless.

## Assumptions
- Deferred work can remain in done/history when it accurately records past implementation, but active queue guidance must point to the v1 core path.

## Acceptance Criteria
1. Engineering active queue names Cairn Core v1 as the batch goal.
2. CocoIndex-specific work is not listed as a required next active story.
3. Existing CocoIndex docs or artifacts are labeled future/reference where needed.
4. Product or planning notes explain why this is a scope reduction, not a feature rejection.
5. A reviewer can tell which stories are required for v1 and which are deferred.

## Validation
- Required checks:
  - Backlog review.
  - Documentation review for active queue and planning artifacts.
- Additional checks:
  - Confirm no required v1 story depends on `deployments/cocoindex-indexer`.

## Dependencies
- `ARCH-20260507-core-v1-indexing-boundary-refresh` is accepted.
- Should follow the docs/runtime/smoke stories so any remaining rich-retrieval references are classified in context.

## Risks
- Historical done items may still contain CocoIndex language; avoid rewriting history unless the current guidance is misleading.

## Open Questions
- Future rich retrieval can stay in engineering intake after v1 stabilizes; no separate epic lane is required yet.

## PM Handoff
- `What changed`: Promoted into active as the fourth Cairn Core v1 story, with deferred release scope preserved.
- `Why it matters`: The backlog needs to stop steering engineers back into the old remote-indexer-first path.
- `Acceptance criteria`: Kept queue clarity and future/reference labeling as the measurable bar.
- `Risks and assumptions`: Preserve history; fix current guidance.
- `Next state recommendation`: Handle after the core docs, runtime, and smoke work reveal which references are still confusing.

## Engineering Handoff
- `Implementation summary`: Added explicit superseded/deferred notes to `docs/adr/ADR-local-development-emulation.md` and `Flywheel/flywheel/artifacts/planning/PLAN-20260506-local-dev-cocoindex-environment.md`. Removed old deployment/prototype folders from the active tree after human confirmation that v1 should shed cruft: `deployments/azure-container-apps-indexer/`, `deployments/local-indexer/`, `deployments/cocoindex-indexer/`, and the compose/Azurite smoke files under `deployments/local-dev/`. Rewrote `deployments/local-dev/README.md` around the no-service `local_fs` smoke. Updated the engineering active queue to show that required Cairn Core v1 stories are complete.
- `Why this is scope reduction, not rejection`: The notes preserve CocoIndex, pgvector, Azurite, and remote indexer material as future rich-retrieval/reference work while making clear they are not required for Cairn Core v1.
- `Validation`: Ran backlog/docs review with `rg` for default/required CocoIndex, remote indexer, Postgres/pgvector, and removed deployment references. Confirmed current required v1 stories do not depend on `deployments/cocoindex-indexer`; historical done items may still mention it as completed history. Re-ran `go test ./...` and `deployments/local-dev/core-smoke.sh` after cleanup.
- `Action and approval notes`: Documentation/backlog edits are `local write`. Destructive cleanup was limited to local cruft and old prototype/deployment artifacts after the human explicitly asked to remove cruft so the project aligns with the new intent. No production actions.
- `Open risks`: Historical done/observer artifacts still contain older CocoIndex language by design; current guidance now marks the path optional/deferred.
- `QA focus areas`: Confirm active queue clarity, deferred labeling on old planning/ADR/prototype docs, and no current required v1 story dependency on `deployments/cocoindex-indexer`.
- `Next state recommendation`: Move to engineering QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed the engineering active queue, current 2026-05-07 required v1 done stories, `ADR-local-development-emulation.md`, the superseded 2026-05-06 CocoIndex planning note, and `deployments/local-dev/README.md`. Current guidance identifies Cairn Core v1 as the active target, removes old prototype/deployment entry points from the active tree, and preserves rich-retrieval thinking as history rather than required v1 scope.
- `Evidence quality call`: Sufficient for backlog/documentation cleanup. QA ran targeted `rg` checks for `deployments/cocoindex-indexer`, CocoIndex, remote indexer, Postgres, and pgvector references across current queues and updated reference docs.
- `Defects`: None.
- `Required fixes`: None.
- `Residual risks`: Older done and observer artifacts still contain historical CocoIndex language, intentionally left as history.
- `Next state recommendation`: Move to engineering done.

## Next Step
- Engineering active queue is clear. Proceed to code review.
