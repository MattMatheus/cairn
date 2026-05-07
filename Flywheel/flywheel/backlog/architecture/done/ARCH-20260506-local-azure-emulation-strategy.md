# Architecture Story: Local Azure Emulation Strategy

## Metadata
- `id`: ARCH-20260506-local-azure-emulation-strategy
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary, ADR-sync-conflict-behavior, ADR-local-development-emulation]
- `decision_owner`: Software Architect
- `success_metric`: One documented local-dev topology decision identifies the blob emulator strategy, auth stance, and service boundaries.

## Decision Scope
- Decide how closely the local development environment should emulate Azure Blob Storage, ACA ingress/auth, and the remote indexer runtime.
- Decide whether to use Azurite directly, a Cairn-specific blob REST shim, or a hybrid.

## Problem Statement
- Cairn needs a local development environment for CocoIndex work that exercises remote sync and remote indexing without requiring Azure.
- The current Azure Blob client expects Azure REST paths and bearer-token auth; that may not be drop-in with a basic Azurite setup.
- The current local indexer scans a mounted filesystem and does not represent the intended blob + pgvector + CocoIndex path.

## Architecture Decision
- Created and accepted `docs/adr/ADR-local-development-emulation.md`.
- Decision: support both `local_fs` remote storage for the fastest local developer loop and Azurite-backed `azure_blob` for full integration testing.
- Decision: use explicit local development auth for local services rather than simulating production Entra/ACA auth in the initial harness.
- Decision: the CocoIndex indexer should read through the configured remote-store backend and write/query Postgres/pgvector while preserving the existing HTTP contract.

## Alternatives Considered
- Azurite only: good Azure fidelity, but too much friction for simple local workflows.
- Cairn-specific blob shim only: easy to tailor, but weak at validating Azure Blob-compatible behavior.
- Real Azure resources: high fidelity, but too slow and expensive for the default developer loop.
- ACA/Entra simulation from the start: deferred because it does not unblock sync/index/CocoIndex integration.

## Risks And Mitigations
- Two storage modes increase implementation surface.
  Mitigation: keep both behind `remotestore.Store` and share conformance tests.
- Azurite auth differs from Azure CLI bearer-token auth.
  Mitigation: make local auth mode explicit and test it independently.
- CocoIndex/model dependencies may slow first run.
  Mitigation: keep the smoke corpus small and document cache/reset behavior.

## Operational Impact
- The local compose stack should include Azurite, Postgres/pgvector, and the indexer service.
- Developer docs should show a quick `local_fs` path and a full Azurite integration path.
- Bicep and ACA deployment docs should be removed or archived because they are not the supported workflow.

## Follow-On Implementation Paths
- `STORY-20260506-remove-azure-bicep-deployment-plan`
- `STORY-20260506-local-dev-compose-harness`
- `STORY-20260506-dev-blob-store-emulator`
- `STORY-20260506-cocoindex-indexer-service`
- `STORY-20260506-developer-quickstart-and-smoke`

## QA Verdict
- Verdict: pass.
- Evidence summary: human review accepted the ADR direction; acceptance criteria are met by `docs/adr/ADR-local-development-emulation.md`.
- Defects: none filed.
- Next state recommendation: done.

## Acceptance Criteria
1. The chosen blob emulation approach is explicit and justified.
2. The local auth stance is explicit and does not require real Azure credentials.
3. Follow-on engineering stories have enough detail to implement the stack without reopening the topology decision.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
