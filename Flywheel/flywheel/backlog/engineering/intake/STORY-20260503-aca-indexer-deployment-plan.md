# Story: Azure Container Apps Indexer Deployment Plan

## Metadata
- `id`: STORY-20260503-aca-indexer-deployment-plan
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-indexing-query-boundary]
- `success_metric`: Cairn has a concrete ACA deployment/auth plan for the CocoIndex-backed remote indexer.
- `release_scope`: optional

## Problem Statement
- The indexing ADR leaves Azure Container Apps packaging and auth enforcement as follow-up design work.

## Scope
- In:
  - Define ACA deployment topology for the indexer service.
  - Define auth enforcement approach compatible with Azure CLI identity from Cairn.
  - Define required Azure resources, secrets boundaries, and environment variables.
  - Identify operational checks and failure modes.
- Out:
  - Implementing Terraform/Bicep.
  - Live Azure deployment.
  - Sync/indexer production runbooks.

## Assumptions
- V1 can document deployable topology before infrastructure automation.

## Acceptance Criteria
1. Deployment plan identifies ACA, storage, identity, and network boundaries.
2. Auth flow is described without storing secrets in Cairn config.
3. Operational checks and failure modes are listed.
4. Follow-up implementation stories are identified.

## Validation
- Required checks:
  - Manual review against indexing ADR and north-star enterprise constraints.

## Dependencies
- `STORY-20260503-cocoindex-contract-prototype`
- `STORY-20260503-azure-blob-sync-adapter`

## Risks
- Avoid premature production hardening before a local packaging prototype exists.

## Open Questions
- Exact ACA auth enforcement mechanism.

## Next Step
- PM should keep this behind local packaging unless cloud deployment becomes the priority.

