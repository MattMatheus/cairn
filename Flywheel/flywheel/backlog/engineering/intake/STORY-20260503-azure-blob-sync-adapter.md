# Story: Azure Blob Sync Adapter Foundation

## Metadata
- `id`: STORY-20260503-azure-blob-sync-adapter
- `owner_role`: Software Architect
- `status`: intake
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
- Which Azure SDK package, if any, should be used for v1.

## Next Step
- PM should refine after read-only sync status exists.
