# Story: CocoIndex Local Packaging Prototype

## Metadata
- `id`: STORY-20260503-cocoindex-local-packaging
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary]
- `success_metric`: Cairn has a local packaging prototype for a CocoIndex-backed indexer service that implements the remote index contract.
- `release_scope`: optional

## Problem Statement
- Cairn has a remote index contract but no runnable local service package for teams to try richer indexing.

## Scope
- In:
  - Define local Docker/Podman packaging shape for the indexer service.
  - Sketch or scaffold service endpoints matching `/index/status`, `/index/refresh`, and `/search`.
  - Document local environment variables and volume mounts.
  - Add smokeable contract tests where practical without requiring embeddings.
- Out:
  - Production ACA deployment.
  - Real embedding pipeline requirement for tests.
  - Hosted SaaS packaging.

## Assumptions
- A stubbed local service can prove endpoint contracts before full CocoIndex pipeline integration.

## Acceptance Criteria
1. Local packaging files or docs exist for an indexer service.
2. Service shape matches the remote index contract.
3. Required env vars and mounts are documented.
4. Tests/smoke guidance do not require paid model calls.

## Validation
- Required checks:
  - Contract smoke or documented manual smoke.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against CocoIndex contract notes.

## Dependencies
- `STORY-20260503-cocoindex-contract-prototype`

## Risks
- Keep this as a prototype; avoid turning packaging into production ops.

## Open Questions
- Whether the indexer service should live in this repo or a sibling package long-term.

## Engineering Handoff
- Implemented 2026-05-03.
- Added `cmd/cairn-indexer` local HTTP indexer prototype.
- Added `internal/remoteindex.Service` with `/index/status`, `/index/refresh`, and `/search`.
- Added deterministic lexical search over managed Cairn markdown so smoke tests require no embeddings or paid model calls.
- Added local Docker/Podman packaging under `deployments/local-indexer/`.
- Documented required environment variables and workspace volume mount.

## QA Handoff
- Accepted 2026-05-03.
- Verified service endpoint shape matches the remote index contract.
- Verified status, refresh, and search through contract smoke tests.
- Verified dry-run refresh does not mark the service fresh.
- Verified packaging docs include env vars, volume mount, build/run commands, and smoke commands.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260503-aca-indexer-deployment-plan`.
