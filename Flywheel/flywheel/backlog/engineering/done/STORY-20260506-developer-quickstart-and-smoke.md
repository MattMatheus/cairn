# Story: Developer Quickstart And Smoke Tests

## Metadata
- `id`: STORY-20260506-developer-quickstart-and-smoke
- `owner_role`: Technical Writer
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260506-local-azure-emulation-strategy, ADR-local-development-emulation]
- `success_metric`: A new developer can complete the documented local smoke path in under 15 minutes after dependencies are installed.
- `release_scope`: required

## Problem Statement
- Developers need an obvious, low-friction path to run Cairn locally with remote sync and CocoIndex indexing enabled.
- Existing docs emphasize CLI local-only use and obsolete Azure deployment planning.

## Scope
- In:
  - Add local development quickstart documentation.
  - Provide sample workspace config for `local_fs`, Azurite, and local indexer.
  - Add a smoke-test script and documented command sequence.
  - Document reset/cleanup commands for local volumes and generated workspace state.
- Out:
  - Production deployment docs.
  - Deep CocoIndex internals documentation.

## Assumptions
- The smoke path should use a throwaway workspace, not the repo root.
- Developers may use Docker or Podman, but one path should be primary.

## Acceptance Criteria
1. Docs show how to start services, initialize a workspace, configure local remote sync/indexing, push, refresh, search, and pull.
2. Docs include expected outputs or success criteria for each smoke step.
3. Docs include troubleshooting for ports, Postgres readiness, Azurite, model downloads, and local auth/token behavior.
4. Obsolete Azure/Bicep setup links are removed or redirected.
5. A smoke script exercises the documented happy path from a throwaway workspace.

## Validation
- Required checks:
  - Follow the quickstart from a fresh temp workspace.
  - Run the smoke script or command sequence end-to-end.
- Additional checks:
  - Confirm docs do not tell developers to use real Azure resources for local development.

## Dependencies
- `STORY-20260506-remove-azure-bicep-deployment-plan`
- `STORY-20260506-local-dev-compose-harness`
- `STORY-20260506-dev-blob-store-emulator`
- `STORY-20260506-cocoindex-indexer-service`

## Risks
- Docs may become stale if the compose harness changes during implementation.

## Open Questions
- Resolved by PM: provide a script first; Make targets can be added later if they fit repo conventions.

## Next Step
- Done.

## Engineering Handoff
- Added `deployments/local-dev/smoke.sh` as the primary developer smoke path.
- Added sample workspace configs:
  - `deployments/local-dev/workspace-config.azurite.yaml`
  - `deployments/local-dev/workspace-config.local_fs.yaml`
- Expanded `deployments/local-dev/README.md` with start, manual smoke, expected outputs, reset commands, config snippets, and troubleshooting.
- Linked the local harness from the root README and user docs.

## Engineering Validation
- `deployments/local-dev/smoke.sh`
- `GOCACHE=$PWD/.cache/go-build go test ./...`
- Documentation scan confirmed local-dev docs do not instruct developers to use real Azure resources for the local loop.

## QA Verdict
- Pass.

## QA Evidence
- Fresh smoke script run passed:
  - started Podman Compose services
  - initialized a throwaway workspace
  - pushed managed markdown to Azurite
  - refreshed the local remote indexer
  - searched through `remote_index.url`
  - pulled a remote create into a second throwaway workspace
- Regression suite passed: `GOCACHE=$PWD/.cache/go-build go test ./...`.
