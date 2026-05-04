# Azure Container Apps Indexer Deployment Plan

This plan describes a deployable Azure Container Apps shape for the Cairn remote indexer contract. A reviewable Bicep skeleton lives in `infra/`.

## Goals

- Host the CocoIndex-backed indexer behind Cairn's stable HTTP contract.
- Keep Cairn config free of stored secrets.
- Use Azure identity and network boundaries that fit per-pod workspaces.
- Preserve local Cairn usefulness when the remote indexer is unavailable.

## Contract

The service exposes:

- `POST /index/status`
- `POST /index/refresh`
- `POST /search`

The local packaging prototype in `deployments/local-indexer/` is the contract smoke target. A production CocoIndex image can replace the prototype as long as these endpoints keep the same request and response shapes.

## Topology

```text
Developer or agent host
  Azure CLI credential
  Cairn CLI / MCP
    |
    | Bearer token from Azure CLI or managed identity flow
    v
Azure Container Apps ingress
  Auth policy validates Microsoft Entra ID token
    |
    v
Container App: cairn-indexer
  CocoIndex pipeline
  Reads pod workspace objects
  Writes derived index artifacts
    |
    +--> Azure Blob Storage: pod workspace documents
    +--> Azure Database for PostgreSQL + pgvector: semantic index
    +--> Azure Log Analytics: logs and traces
```

## Azure Resources

- Resource group per environment, not necessarily per pod.
- Azure Container Apps environment.
- Container App for the indexer service.
- Azure Container Registry or GitHub Container Registry for the image.
- Azure Blob Storage account/container for pod workspace documents.
- Azure Database for PostgreSQL Flexible Server with pgvector for derived index artifacts.
- User-assigned managed identity for the Container App.
- Log Analytics workspace for ACA logs.
- Optional private endpoints for Storage and PostgreSQL when enterprise network policy requires them.
- Optional Azure Key Vault for infrastructure-managed secrets such as PostgreSQL connection material.

## Identity And Auth

Cairn callers should not store API keys in workspace config.

Preferred flow:

1. User or agent signs in with Azure CLI.
2. Cairn obtains a bearer token for the configured indexer audience.
3. ACA ingress validates Microsoft Entra ID tokens before forwarding traffic.
4. The indexer authorizes the workspace by comparing token claims to allowed pod/workspace configuration.
5. The indexer uses its managed identity to read Blob Storage and reach backing services.

Fallback if ACA built-in auth is insufficient:

- Keep ACA ingress private or authenticated at the edge.
- Add token validation middleware in the indexer service.
- Accept only Microsoft Entra ID bearer tokens.
- Validate issuer, audience, expiry, tenant, and allowed group/app claims.

Do not:

- Store indexer bearer tokens in `.cairn/config.yaml`.
- Put Azure storage account keys in Cairn workspace files.
- Share one broad identity across all pods without claim or config scoping.

## Secrets Boundary

Cairn workspace config may contain:

- Indexer URL.
- Azure tenant id.
- Token audience or app id URI.
- Workspace id.
- Storage container names or logical pod ids.

Cairn workspace config must not contain:

- Client secrets.
- Storage account keys.
- PostgreSQL passwords.
- Long-lived bearer tokens.

Indexer/runtime secrets are owned by Azure infrastructure:

- Managed identity assignments.
- Key Vault references if PostgreSQL requires password auth.
- ACA environment secret references.

## Environment Variables

Container App:

- `CAIRN_INDEXER_WORKSPACE_ID`: pod workspace id served by this app or shard.
- `CAIRN_STORAGE_ACCOUNT`: Azure Storage account name.
- `CAIRN_STORAGE_CONTAINER`: Blob container containing workspace objects.
- `CAIRN_STORAGE_PREFIX`: optional pod prefix within the container.
- `CAIRN_POSTGRES_DSN`: PostgreSQL connection string or Key Vault-backed secret reference.
- `CAIRN_AUTH_AUDIENCE`: expected token audience for direct token validation.
- `CAIRN_AUTH_TENANT_ID`: expected Entra tenant id.
- `CAIRN_LOG_LEVEL`: default `info`.

Cairn client config:

- `remote_index.url`
- `remote_index.tenant_id`
- `remote_index.audience`
- `workspace_id`

## Network Boundaries

Minimum viable:

- ACA external ingress with Entra auth.
- Storage and PostgreSQL reachable from ACA.
- No public direct database access.

Enterprise preferred:

- ACA environment integrated with VNet.
- Private endpoint for Storage.
- Private endpoint for PostgreSQL.
- Public ingress limited to authenticated HTTPS.
- Restrict egress to Storage, PostgreSQL, Azure Monitor, and identity endpoints.

## Operational Checks

- `POST /index/status` returns `available: true`.
- Status reports freshness and last refresh time after successful refresh.
- `POST /index/refresh` returns either `accepted: true` with a job id or `refreshed: true` for synchronous completion.
- `POST /search` returns stable Cairn result shape.
- ACA revision is healthy and has at least one ready replica.
- Container logs show no auth validation failures for valid users.
- Storage access uses managed identity, not account keys.
- PostgreSQL extension `vector` is installed when semantic storage is enabled.
- Search degradation in Cairn reports remote unavailable warnings when ACA is down or auth fails.

## Failure Modes

- Auth token missing or invalid: return `401`; Cairn reports remote auth unavailable.
- Auth token valid but workspace unauthorized: return `403`; Cairn reports remote service/auth warning with no local failure.
- Storage unavailable: refresh returns retryable service failure; search may serve stale index if available.
- PostgreSQL unavailable: status returns unavailable or stale; refresh fails retryably.
- Index refresh accepted but not complete: response includes `accepted: true`, `job_id`, and `refreshed: false`.
- Index schema mismatch: status reports unavailable with operator-facing message.
- ACA cold start or scale to zero: Cairn degrades gracefully and suggests retry/status.

## Follow-Up Stories

- Add Entra ID app registration and token validation middleware.
- Add Cairn config schema fields for `remote_index`.
- Add Azure CLI token provider for `remoteindex.HTTPClient`.
- Add remote index status/refresh CLI configuration wiring.
- Add CocoIndex pipeline image that writes Postgres/pgvector artifacts.
- Add production runbook for refresh failures and index freshness checks.

## Infrastructure Skeleton

The `infra/main.bicep` module is a scaffold for review and local rendering. It models:

- Log Analytics workspace.
- Azure Container Apps environment.
- User-assigned managed identity.
- Container App with external HTTPS ingress.
- Storage Blob Data Reader assignment for an existing workspace storage account.
- PostgreSQL connection as an infrastructure-owned secret reference name, not a checked-in value.

Render the template locally:

```sh
az bicep build --file deployments/azure-container-apps-indexer/infra/main.bicep
```

Preview a deployment:

```sh
az deployment group what-if \
  --resource-group <resource-group> \
  --template-file deployments/azure-container-apps-indexer/infra/main.bicep \
  --parameters @deployments/azure-container-apps-indexer/infra/main.parameters.example.json
```

The example parameters file contains no secrets. Replace placeholder IDs and image names before any real deployment.
