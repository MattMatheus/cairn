# Story: Azure Blob Sync Adapter Foundation

## Metadata
- `id`: STORY-20260503-azure-blob-sync-adapter
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-sync-conflict-behavior]
- `success_metric`: Cairn has an Azure Blob adapter boundary for remote manifest/file IO using Azure CLI identity without implementing unsafe merge behavior.
- `release_scope`: required

## Problem Statement
- Sync safety primitives are local-only. V1 needs an Azure Blob remote adapter boundary before `sync_pull` and `sync_push` can be implemented.

## Scope
- In:
  - Define remote storage adapter interface for manifest and workspace object IO.
  - Implement Azure Blob adapter skeleton using configured account/container/prefix.
  - Use Azure CLI identity/token acquisition boundary without storing secrets.
  - Add tests around URL/path mapping and adapter behavior with fakes.
- Out:
  - Production Azure integration tests.
  - Sync pull/push mutation orchestration.
  - Conflict merge.

## Assumptions
- Networked Azure calls can be isolated behind interface and fakes for unit tests.

## Acceptance Criteria
1. Adapter interface supports remote manifest read/write and object read/write/list.
2. Azure path mapping mirrors workspace root under optional prefix.
3. Credential boundary uses Azure CLI identity assumptions without secrets in config.
4. Tests cover path mapping and fake adapter contract.

## Validation
- Required checks:
  - Unit tests for adapter contract/path mapping.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against sync ADR and north-star auth section.

## Dependencies
- `STORY-20260503-sync-status-conflict-report`
- `STORY-20260503-workspace-init-config`

## Risks
- Avoid live Azure requirements in unit tests.

## Open Questions
- Resolved for foundation: use a standard-library HTTP adapter boundary with an injectable Azure CLI token provider; defer Azure SDK adoption until live integration needs it.

## Next Step
- Engineering should implement the remote adapter boundary and unit tests without live Azure dependencies.

## PM Handoff
- Promoted on 2026-05-03 after read-only sync status landed.
- Keep this as an adapter foundation only: no sync pull/push orchestration or merge behavior.
- Unit tests should use fakes/httptest and must not require Azure credentials or network access.

## Engineering Handoff
- Added `internal/remotestore.Store` for remote manifest and workspace object IO.
- Added `MemoryStore` fake implementing manifest read/write, object read/write, and object listing.
- Added Azure Blob HTTP adapter skeleton with account/container/prefix config, workspace-root path mapping, object read/write/list, and manifest read/write.
- Added Azure CLI token provider boundary using `az account get-access-token` for storage tokens without storing secrets.
- Updated generated workspace config to include a `pod-remote.account` placeholder.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/remotestore`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed adapter interface supports remote manifest read/write and object read/write/list.
- Confirmed Azure Blob path mapping mirrors the workspace root under optional prefix.
- Confirmed Azure credential boundary uses Azure CLI token acquisition and does not store secrets.
- Confirmed tests use in-memory/fake HTTP transport only; no live Azure or local network listener is required.
- QA polish applied before acceptance: added explicit Azure manifest read/write coverage for `.cairn/remote-manifest.json`.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote the read-only MCP server transport story so existing operation adapters can be exposed through a local MCP process.
