# Architecture Story: Sync And Conflict Behavior ADR

## Metadata
- `id`: ARCH-20260502-sync-conflict-behavior
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-sync-conflict-behavior]
- `decision_owner`: Software Architect
- `success_metric`: A reviewer can implement Azure Blob sync, manifests, local sync state, and refused-conflict behavior without designing merge semantics.

## Decision Scope
- Define Cairn v1 on-demand sync behavior for Azure Blob, including remote layout, ignore rules, manifest schema, local state, base detection, changed object semantics, refused conflicts, archive/move handling, and operation permissions.

## Problem Statement
- The north-star says Cairn sync refuses conflicts rather than merging. That is a good v1 constraint, but implementation still needs exact manifest/state semantics so a refused sync never corrupts local or remote knowledge.

## Inputs
- Existing decisions:
  - `docs/product/north-star.md`
- Existing architecture artifacts:
  - None yet.
- Constraints:
  - Azure Blob is the first shared backend.
  - Each pod should have its own Azure Storage Account.
  - Cloud drive sync and OneDrive are out of scope.
  - Sync is on-demand through CLI or MCP.
  - Local sync state lives at `/.cairn/sync-state.json`.
  - Remote manifest lives at `/.cairn/remote-manifest.json`.
  - The last accepted remote manifest is the divergence base.
  - Refused sync must not update local state, overwrite files, overwrite remote files, or publish a new manifest.
  - Git is not the backend; only light Git nudging is in scope.

## Outputs Required
- Decision updates:
  - `docs/adr/ADR-sync-conflict-behavior.md` defining sync push/pull/status behavior, manifest shape, local state shape, ignored path handling, conflict detection, and refusal behavior.
- Architecture artifacts:
  - Manifest field table.
  - Local state field table.
  - Push and pull sequence sketches.
  - Conflict examples for content edit, create, move, archive, and delete.
- Risks and tradeoffs:
  - Refusing conflicts avoids bad merges but requires clear recovery guidance.
  - Whole-workspace sync is simpler but can be heavy for larger pods.
  - Delete semantics are risky when purge is CLI-only and archive is preferred.

## Alternatives Considered
- Git-backed sync.
- Cloud-drive sync.
- Automatic three-way merge.
- Path-only manifests without document ids.
- Per-document sync rather than workspace sync.

## Operational Impact
- Defines how pods share knowledge safely.
- Defines what MCP sync tools are allowed to mutate.
- Determines recovery paths when two users or agents diverge.

## Acceptance Criteria
1. ADR defines the v1 Azure Blob layout and prefix behavior.
2. ADR defines remote manifest and local sync state schemas at decision level.
3. ADR defines how changes are detected for creates, edits, moves, archives, and deletes.
4. ADR defines exactly when sync refuses and what state remains unchanged.
5. ADR defines user-facing recovery guidance after refusal.

## Review Focus
- Confirm conflict refusal is deterministic and explainable.
- Confirm purge/delete constraints are respected.
- Confirm no Git backend assumptions leak into the sync design.

## Next Step
- Architecture QA should review `docs/adr/ADR-sync-conflict-behavior.md`.

## Architecture Handoff
- `Architecture decision`: Drafted `docs/adr/ADR-sync-conflict-behavior.md` as a proposed ADR covering Azure Blob layout, manifests, local sync state, change detection, conflict refusal, and recovery guidance.
- `Alternatives considered`: Git-backed sync, cloud-drive sync, automatic three-way merge, per-document sync, and path-only manifests.
- `Key risks`: Refused conflicts are safe but manual; manifest schemas become compatibility surfaces; delete/archive semantics need careful tests.
- `Follow-on implementation paths`: `.cairnignore`, manifest generation, local sync state, sync status/pull/push, conflict reporting, sync tests.
- `Next state recommendation`: Move to architecture QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: ADR was reviewed against the north-star document and the story acceptance criteria. It covers Azure Blob layout, manifest and local state fields, change detection, conflict refusal semantics, recovery guidance, and the non-Git backend constraint.
- `Evidence quality call`: Sufficient for architecture acceptance; exact manifest JSON schema can be refined during implementation without splitting this parent ADR.
- `Defects`: None.
- `Required fixes`: None.
- `Split decision`: No split required now. Blob manifest schema may become a child ADR only if implementation reveals compatibility choices that need a separate durable decision.
- `Completed work summary`: Accepted the sync and conflict behavior ADR.
- `Next suggested or required step`: Use this ADR to seed implementation stories for `.cairnignore`, manifest/state handling, sync status/pull/push, and conflict tests.
- `Next state recommendation`: Move to architecture done.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
