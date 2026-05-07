# Planning: Local Development Environment For CocoIndex Integration

## Status

Superseded for the Cairn Core v1 default path by `PLAN-20260507-cairn-core-v1-rescope` and `ARCH-20260507-core-v1-indexing-boundary-refresh`.

This plan remains useful reference material for future rich-retrieval work. CocoIndex, Postgres/pgvector, Azurite, and the remote indexer are not required v1 infrastructure.

## Objective

Make Cairn easy to run on a developer machine while developing and testing the CocoIndex-backed remote index path. The local setup should simulate the Azure runtime shape closely enough to exercise remote sync, remote indexing, and search without requiring an Azure subscription.

## Current Findings

- Cairn already has a Go CLI/MCP surface and local metadata index.
- The current local indexer container in `deployments/local-indexer/` is a contract smoke target. It scans a mounted filesystem path and does not use Blob Storage, Postgres, pgvector, or CocoIndex.
- Remote sync is implemented against an Azure Blob REST shape through `internal/remotestore/azure_blob.go`.
- The Azure Container Apps Bicep skeleton and deployment plan are obsolete for the intended workflow and can be removed.
- The planned production topology still gives useful local-dev requirements: an indexer HTTP service, blob-backed workspace objects, and Postgres/pgvector for semantic index state.

## Desired Local Topology

Developer host:
- runs `cairn` CLI/MCP against a local workspace
- uses local config values only, with no checked-in secrets

Local containers:
- Blob-compatible workspace object store
- Postgres with pgvector
- CocoIndex-backed Cairn indexer service exposing:
  - `POST /index/status`
  - `POST /index/refresh`
  - `POST /search`

Optional later container:
- local auth/ingress shim that simulates ACA-authenticated requests and forwarded principal headers

## Scope Boundary

In scope:
- Remove unused Azure Bicep/deployment artifacts.
- Add a local development harness under the repo.
- Provide a repeatable developer quickstart and smoke-test path.
- Add local/dev configuration support where needed for blob endpoint, static dev token behavior, and indexer URL wiring.
- Replace or supplement the prototype indexer with a CocoIndex-backed service that uses Blob input and pgvector output.

Out of scope for this batch:
- Real Azure deployment automation.
- Production Entra validation.
- Production Container Apps configuration.
- Key Vault, private endpoints, Log Analytics, or managed identity setup.
- Full hosted multi-pod tenancy.

## Proposed Work Items

1. `ARCH-20260506-local-azure-emulation-strategy`
2. `STORY-20260506-remove-azure-bicep-deployment-plan`
3. `STORY-20260506-local-dev-compose-harness`
4. `STORY-20260506-dev-blob-store-emulator`
5. `STORY-20260506-cocoindex-indexer-service`
6. `STORY-20260506-developer-quickstart-and-smoke`

## Recommended Order

1. Decide the local Azure emulation boundary.
2. Remove obsolete Azure deployment docs/code so developers do not follow a dead path.
3. Add the compose harness with Postgres/pgvector and placeholders for blob/indexer services.
4. Implement or wire the blob emulator path and Cairn dev auth behavior.
5. Add the CocoIndex indexer service.
6. Finish docs and smoke scripts that prove the complete local loop.

## Success Signals

- A developer can start the local stack with one documented command.
- A fresh Cairn workspace can push/pull against the local blob store.
- The local indexer can refresh from blob-backed workspace documents.
- Search returns results from the CocoIndex/pgvector-backed remote index path.
- The docs clearly distinguish local development from production Azure deployment.

## Risks

- Azurite may not match the current bearer-token-oriented Azure Blob client without extra configuration or code changes.
- CocoIndex packaging may pull larger Python/model dependencies than the Go prototype container.
- Initial embedding model downloads can make first-run setup slow unless documented or cached.
- The current indexer contract is stable, but internal indexer implementation language/runtime may shift from Go-only to Python or mixed Go/Python.

## Assumptions

- Local developer ergonomics matter more than exact Azure parity for the next integration step.
- A static dev token or no-op local auth path is acceptable for local-only testing.
- Postgres/pgvector is the right first semantic index target because it matches the existing CocoIndex references and Azure plan.
- The Bicep deployment skeleton will not be used.

## Next Stage Recommendation

Next stage: `architect` for `ARCH-20260506-local-azure-emulation-strategy`.

After that decision lands, move the engineering stories through PM refinement and implement them in the recommended order.
