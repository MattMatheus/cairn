# Story: Remove Azure Bicep Deployment Plan

## Metadata
- `id`: STORY-20260506-remove-azure-bicep-deployment-plan
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-local-development-emulation]
- `success_metric`: No developer-facing docs point to Bicep or Azure Container Apps deployment as the supported path.
- `release_scope`: required

## Problem Statement
- The repository contains Azure Container Apps Bicep scaffolding and deployment docs that will not be used.
- Keeping them beside the new local development path creates confusion for developers trying to run Cairn locally.

## Scope
- In:
  - Remove obsolete Bicep files and parameters.
  - Remove or rewrite docs that present ACA deployment as the active plan.
  - Keep product notes that remain useful for runtime shape only if they are clearly marked historical or replaced.
- Out:
  - Any production Azure deployment replacement.
  - Changes to sync/indexer runtime behavior.

## Assumptions
- The Bicep code is not needed for future deployments.
- Local development is the next priority.

## Acceptance Criteria
1. `deployments/azure-container-apps-indexer/infra/` and checked-in Bicep parameter examples are removed.
2. README and user docs no longer direct developers toward Bicep, ACA, or real Azure setup for local development.
3. Remaining product notes either point to the local development harness or are clearly marked historical/deferred.
4. References to the accepted local development ADR are present where they replace the old Azure deployment direction.

## Validation
- Required checks:
  - `rg -n "bicep|Container Apps|ACA" README.md docs deployments`
  - Manual doc review for stale Azure deployment instructions.
- Additional checks:
  - Confirm no code imports or tests depend on removed deployment files.

## Dependencies
- None.

## Risks
- Removing too much historical context could obscure why the runtime still uses Azure-shaped terms such as blob and remote index.

## Open Questions
- Resolved by PM: delete active Bicep/deployment scaffolding; keep only product notes that are still useful and clearly marked historical/deferred.

## Next Step
- QA: verify obsolete Azure/Bicep deployment artifacts are gone, developer-facing docs point at local development, and remaining Azure/ACA/Bicep references are historical or deferred context only.

## Engineering Handoff
- Action model: local write.
- Approval: not required; no risky, sensitive, or production action taken.
- Changed implementation/docs:
  - Removed `deployments/azure-container-apps-indexer/` deployment scaffolding.
  - Updated `README.md` and `docs/user/README.md` to point at `docs/adr/ADR-local-development-emulation.md`.
  - Updated product/ADR notes to describe production auth/deployment as historical or deferred instead of the active local development path.
- Validation:
  - Ran `rg -n "bicep|Bicep|Container Apps|ACA|azure-container-apps-indexer|INFRA-PREP|main\\.parameters|main\\.bicep" README.md docs deployments`.
  - Ran `find deployments -maxdepth 4 -type f | sort`.
  - Ran `rg -n "azure-container-apps-indexer|INFRA-PREP|main\\.bicep|main\\.parameters|Container Apps|ACA" --glob '!Flywheel/**' .`.
- Validation result:
  - Active deployment files are removed.
  - Remaining matches are in ADR/product notes and are framed as historical or deferred production auth context.
  - `deployments/` now contains only `deployments/local-indexer/` files.
- Open risks:
  - Future implementation stories still need to add the replacement `deployments/local-dev/` path.
- QA focus:
  - Confirm README/user docs no longer direct developers toward real Azure setup.
  - Confirm no checked-in Bicep or parameter examples remain.
  - Confirm retained Azure/ACA references are not active setup instructions.

## QA Verdict
- Verdict: pass.
- Evidence summary:
  - `find deployments -path '*azure-container-apps-indexer*' -print` returned no paths after cleanup.
  - `find deployments -maxdepth 4 -type f | sort` lists only `deployments/local-indexer/Dockerfile` and `deployments/local-indexer/README.md`.
  - `rg -n "azure-container-apps-indexer|INFRA-PREP|main\\.bicep|main\\.parameters" --glob '!Flywheel/**' .` returned no matches.
  - `rg -n "bicep|Bicep|Container Apps|ACA|azure-container-apps-indexer|INFRA-PREP|main\\.parameters|main\\.bicep" README.md docs deployments` returns only historical/deferred context in ADR/product notes, not active developer setup instructions.
- Defects: none filed.
- Required fixes: removed empty `deployments/azure-container-apps-indexer/` directories found during QA.
- Next state recommendation: done.
