# Story: Local Dev Compose Harness

## Metadata
- `id`: STORY-20260506-local-dev-compose-harness
- `owner_role`: SRE
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260506-local-azure-emulation-strategy, ADR-local-development-emulation]
- `success_metric`: A developer can start required local services with one documented command.
- `release_scope`: required

## Problem Statement
- Developers need a repeatable local environment for Cairn remote sync and CocoIndex integration.
- Today only the prototype local indexer container exists; Postgres/pgvector and blob emulation are not packaged as a single development stack.

## Scope
- In:
  - Add a local development deployment folder.
  - Add compose configuration for Postgres/pgvector, Azurite, and indexer service.
  - Define local environment variables and ports.
  - Provide health checks where practical.
- Out:
  - Production deployment automation.
  - Real Azure credentials or cloud resources.

## Assumptions
- Docker Compose or Podman Compose is acceptable for the first local harness.
- Postgres should use a pgvector image.

## Acceptance Criteria
1. `deployments/local-dev/` contains the local harness entrypoint.
2. The local stack definition includes Postgres/pgvector, Azurite, and an indexer service target that can initially point at the current contract fixture.
3. The stack uses deterministic local ports and documented environment variables.
4. The stack can be started, inspected, and stopped without mutating repository source files.

## Validation
- Required checks:
  - Compose config validates with the chosen compose tool.
  - Postgres health check reaches ready state.
  - Indexer status endpoint is reachable when implemented.
- Additional checks:
  - Confirm volumes are named and safe to remove/reset.

## Dependencies
- `ADR-local-development-emulation`
- `STORY-20260506-remove-azure-bicep-deployment-plan`

## Risks
- Container startup may be slow if model downloads happen inside the indexer path.
- Podman and Docker compatibility may diverge.

## Open Questions
- Resolved by PM: use `deployments/local-dev/`.

## Next Step
- QA: verify the compose harness shape, deterministic ports/env docs, named volumes, and validation evidence.

## Engineering Handoff
- Action model: local write.
- Approval: not required; no risky, sensitive, or production action taken.
- Changed implementation/docs:
  - Added `deployments/local-dev/compose.yml`.
  - Added `deployments/local-dev/.env.example`.
  - Added `deployments/local-dev/README.md`.
- Validation:
  - `ruby -e 'require "yaml"; data = YAML.load_file("deployments/local-dev/compose.yml"); abort("missing services") unless data["services"].keys.sort == %w[azurite indexer postgres]; puts data["services"].keys.sort.join(",")'`
  - `rg -n "postgres|azurite|indexer|CAIRN_INDEXER_PORT|CAIRN_AZURITE_BLOB_PORT|CAIRN_POSTGRES_PORT|postgres_data|azurite_data" deployments/local-dev`
  - Attempted `docker compose -f deployments/local-dev/compose.yml --env-file deployments/local-dev/.env.example config`; Docker is not installed in this environment.
  - Attempted `podman compose -f deployments/local-dev/compose.yml --env-file deployments/local-dev/.env.example config`; Podman socket access is blocked in this environment.
- Validation result:
  - YAML parses and contains the expected `azurite`, `indexer`, and `postgres` services.
  - Harness docs and env example document deterministic ports and service settings.
  - Runtime compose validation and service health checks could not be executed because no usable container runtime is available in this sandbox.
- Open risks:
  - QA should run compose validation on a developer machine with Docker or Podman.
- QA focus:
  - Confirm compose file syntax with real Docker/Podman.
  - Confirm `postgres_data` and `azurite_data` are named volumes and resettable with `down -v`.
  - Confirm indexer service can use the current Go fixture as the initial target.

## QA Verdict
- Verdict: pass.
- Evidence summary:
  - Static YAML shape check passed: services include `postgres`, `azurite`, and `indexer`; volumes include `postgres_data` and `azurite_data`.
  - Docs/env scan confirmed deterministic ports, environment variables, compose start/inspect/stop commands, and `down -v` reset guidance.
  - After Codex permissions were updated, `podman compose -f deployments/local-dev/compose.yml --env-file deployments/local-dev/.env.example config` passed.
  - `podman compose -f deployments/local-dev/compose.yml --env-file deployments/local-dev/.env.example up --build -d` built and started `postgres`, `azurite`, and `indexer`.
  - `podman compose ... ps` showed all three services running with expected host ports.
  - `podman exec local-dev-postgres-1 pg_isready -U cocoindex -d cocoindex` passed.
  - `CREATE EXTENSION IF NOT EXISTS vector` succeeded and `pg_extension` reported `vector`.
  - `curl http://localhost:8080/index/status` returned `available: true`.
- Defects: none filed.
- Required fixes: none.
- Residual risk: the current indexer service is still the Go contract fixture until the CocoIndex-backed service lands.
- Next state recommendation: done.
