# ADR: Local Development Emulation

## Status

Historical. Superseded for the Cairn Core v1 default path by `ADR-indexing-query-boundary`, refreshed on 2026-05-07.

## Current V1 Note

Cairn Core v1 does not include this full local service stack. The default v1 path is local markdown files, local SQLite metadata/full-text search, MCP, and `remote_sync.provider: local_fs` for no-service blob-style sync. Azurite, Postgres/pgvector, CocoIndex, and the remote indexer are future rich-retrieval or Azure-emulation work, not active local development infrastructure.

## Context

Earlier planning explored a full local development environment for remote sync and CocoIndex-backed indexing. That environment was meant to let developers run the tool on their own machines without an Azure subscription, while still exercising the same broad shape expected in hosted operation:

- workspace objects live behind a remote-store boundary
- Cairn CLI/MCP can push, pull, validate, refresh, and search
- the remote indexer exposes the stable Cairn HTTP contract
- CocoIndex writes semantic artifacts to Postgres with pgvector

The Azure Container Apps Bicep skeleton, local indexer container, CocoIndex service prototype, and compose smoke were removed from the active tree during the Cairn Core v1 re-scope. This ADR remains as historical context for future rich retrieval, not as current setup guidance.

Cairn should also accommodate local storage directly. A developer should not need cloud credentials or Azure-specific setup just to work on document lifecycle, sync behavior, or indexer integration.

## Decision

Cairn local development supports the `local_fs` remote-store mode for v1. Future rich-retrieval or Azure-emulation work may reintroduce two remote-store modes:

1. `local_fs`: a first-class local object-store backend rooted at a developer-controlled directory.
2. `azure_blob`: an Azure Blob-compatible backend that can target real Azure later and Azurite locally.

If a full integration harness is reintroduced later, it may use Azurite for blob emulation because it preserves Azure Blob path/list/object semantics better than a Cairn-specific shim. The active quickstart uses `local_fs`.

The local CocoIndex development stack should be:

```text
Developer host
  cairn CLI / MCP
    |
    +--> local workspace files
    +--> local remote store endpoint or directory
    +--> local indexer HTTP endpoint

Compose services
  azurite
    stores workspace objects and .cairn/remote-manifest.json

  postgres
    pgvector-enabled Postgres for CocoIndex target tables

  cairn-indexer
    exposes POST /index/status, POST /index/refresh, POST /search
    reads workspace objects through the configured remote-store backend
    writes/query semantic artifacts through Postgres/pgvector
```

The stable remote indexer HTTP contract remains unchanged:

| Endpoint | Responsibility |
| --- | --- |
| `/index/status` | Report index availability, freshness, and last refresh state |
| `/index/refresh` | Refresh or enqueue refresh for the configured workspace |
| `/search` | Return Cairn-shaped semantic or richer search results |

Local auth will be explicit development auth, not simulated production Entra auth. Local configuration may use one of:

- no auth for loopback-only services
- a static development token
- Azurite development credentials

Production Entra validation, ACA auth headers, managed identity, and private networking are out of scope for the local development harness.

## Config Contract

Active local filesystem mode:

```yaml
remote_sync:
  provider: local_fs
  root: .cairn/local-remote
```

Historical/future Azurite mode:

```yaml
remote_sync:
  provider: azure_blob
  endpoint: http://localhost:10000/devstoreaccount1
  container: cairn
  prefix: pod-a
  auth_mode: azurite

remote_index:
  url: http://localhost:8080
```

The exact field names may be adjusted during implementation, but the architecture requires these capabilities:

- select a remote-store provider
- configure a local filesystem root or Blob endpoint/container/prefix
- avoid Azure CLI login for local development
- point Cairn at a local indexer URL
- configure the indexer with the same store settings and a Postgres/pgvector DSN

## Alternatives Considered

- Azurite only. Rejected as the only path because it keeps local development coupled to Azure-shaped setup even when a developer only needs Cairn storage semantics.
- Cairn-specific blob REST shim only. Rejected because it would be fast to implement but weaker at exercising Azure Blob-compatible behavior.
- Real Azure resources for integration development. Rejected because it makes the core developer loop slower, more expensive, and harder to reproduce.
- Simulate ACA and Entra locally from the start. Deferred because it is not needed to validate sync/indexing behavior and would distract from the CocoIndex integration.

## Consequences

The remote-store boundary becomes more explicit. `azure_blob` is no longer the only remote-store implementation, even if it remains the hosted target.

The local dev harness can serve two developer needs:

- fast local-only workflows through `local_fs`
- Azure-shaped integration workflows through Azurite

The Azure Blob adapter needs local auth support beyond Azure CLI bearer tokens. Azurite support should be implemented deliberately rather than hidden behind production auth code.

The CocoIndex indexer should read through the same store boundary as sync, not by mounting arbitrary local workspace paths as its primary integration mode.

## Operational Impact

Developers should be able to run one documented local stack with Docker Compose or Podman Compose. The stack should expose deterministic ports for Azurite, Postgres, and the indexer.

Generated local state and test workspaces should live outside tracked source paths by default. Docs should show reset commands for compose volumes and local remote-store directories.

Production-oriented Azure deployment docs and Bicep files should be removed or archived so the local development path is not confused with an abandoned deployment path.

## Follow-On Implementation Paths

- Remove obsolete Azure Container Apps Bicep and deployment docs.
- Add a `local_fs` remote-store implementation.
- Extend `azure_blob` configuration to support Azurite/local development auth.
- Add a local compose harness with Azurite, Postgres/pgvector, and the indexer.
- Replace or supplement the Go prototype indexer with a CocoIndex-backed service that reads through the configured remote-store backend.
- Add developer quickstart and smoke tests for both `local_fs` and Azurite-backed flows.

## Risks And Mitigations

- Risk: two storage modes increase implementation surface.
  Mitigation: keep both behind the existing `remotestore.Store` interface and share sync conformance tests.

- Risk: Azurite auth differs from Azure CLI bearer-token auth.
  Mitigation: make local auth mode explicit in config and test it separately from production token acquisition.

- Risk: first-run CocoIndex dependencies and embedding model downloads are slow.
  Mitigation: document caching behavior and keep the smoke corpus small.

- Risk: local filesystem mode hides Blob-specific behavior.
  Mitigation: make Azurite the required full integration path and local filesystem the quick/inner-loop path.
