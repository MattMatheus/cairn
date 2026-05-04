# Story: ACA Indexer Infra Module Skeleton

## Metadata
- `id`: STORY-20260504-aca-indexer-infra-module-skeleton
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary]
- `success_metric`: The documented ACA indexer plan has a checked-in infrastructure module skeleton that can be reviewed before live deployment.
- `release_scope`: optional

## Problem Statement
- The deployment plan identifies ACA resources and follow-up automation, but there is no module scaffold for reviewable infrastructure work.

## Scope
- In:
  - Add a Bicep or Terraform skeleton for ACA indexer resources.
  - Model Container App, Container Apps environment, managed identity, storage access, Log Analytics, and required parameters.
  - Document how to validate/render the module locally.
  - Keep secrets out of checked-in files.
- Out:
  - Live Azure deployment.
  - Production image publishing.
  - Full private networking hardening.

## Assumptions
- Bicep is likely the smallest Azure-native starting point unless the repo establishes Terraform first.

## Acceptance Criteria
1. Infra skeleton exists under `deployments/azure-container-apps-indexer/`.
2. Parameters avoid secret values.
3. README explains render/validation commands.
4. Module reflects the existing deployment plan.

## Validation
- Required checks:
  - Static review against the deployment README.
  - Bicep/Terraform formatting or validation if tooling is available locally.

## Dependencies
- `STORY-20260503-aca-indexer-deployment-plan`

## Risks
- Avoid implying production readiness before auth enforcement is resolved.

## Open Questions
- Resolved for V1 skeleton: use Bicep as the smallest Azure-native scaffold because the repo has no Terraform convention yet.

## Next Step
- Promote `STORY-20260504-remote-indexer-auth-enforcement-spike`.

## Handoff Notes
- Engineering completed 2026-05-04.
- Chose Bicep for the V1 infrastructure skeleton.
- Added `deployments/azure-container-apps-indexer/infra/main.bicep`.
- Added secret-free example parameters.
- Updated the ACA deployment README with render and what-if commands.
- QA completed 2026-05-04 with static review, JSON parameter parsing, and full `go test ./...`.
- Local Azure CLI/Bicep tooling was unavailable, so `az bicep build` was not run.
