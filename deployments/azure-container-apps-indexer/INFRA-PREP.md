# Cairn Remote Infra Prep

This is the operator checklist for preparing Azure infrastructure for Cairn remote sync and the remote indexer. It is written for setup planning, not as a copy-paste production runbook.

## Decisions To Make First

- Azure subscription and tenant.
- Resource group naming.
- Region.
- One storage account per pod or per environment.
- Container naming and optional blob prefix per pod.
- Container registry: ACR or GHCR.
- PostgreSQL/pgvector timing: now for semantic indexing, or later for contract-only smoke.
- Network posture: public authenticated ACA ingress first, or VNet/private endpoints from day one.
- Auth posture: ACA built-in Microsoft Entra auth first, with indexer-side workspace authorization.

## Azure Resources

Minimum viable:

- Resource group.
- Azure Storage account.
- Blob container for workspace files.
- Azure Container Apps environment.
- Container App for the indexer.
- User-assigned managed identity for the indexer.
- Log Analytics workspace.
- Container registry or image source.

Remote semantic indexing path:

- Azure Database for PostgreSQL Flexible Server.
- `vector` extension enabled.
- Database credentials held by Azure infrastructure, not Cairn config.
- Optional Key Vault or ACA secret reference for the PostgreSQL DSN.

Enterprise hardening:

- VNet-integrated ACA environment.
- Private endpoint for Storage.
- Private endpoint for PostgreSQL.
- Restricted egress.
- Azure Monitor alerts for auth failures, refresh failures, and stale index freshness.

## Identity And Permissions

For the indexer managed identity:

- Grant Storage Blob Data Reader on the workspace storage account or container.
- Add write permissions only if the production indexer stores derived artifacts in Blob.
- Grant PostgreSQL access through the chosen auth model.

For users/agents:

- Confirm Azure CLI login works in the target tenant.
- Confirm the token audience that Cairn will request for `remote_index.audience`.
- Decide which users, groups, service principals, or app roles can access each pod workspace.

## Entra App / ACA Auth

Prepare:

- App registration or ACA auth-created app.
- App ID URI or audience, for example `api://cairn-indexer`.
- Tenant id.
- Redirect URI for ACA auth if using browser auth flows.
- ACA auth configured to require authentication and return `401` for unauthenticated API calls.

The current recommendation is in [remote indexer auth enforcement](../../docs/product/remote-indexer-auth-enforcement.md): use ACA built-in Microsoft Entra auth at the edge, then let the indexer authorize workspace access from authenticated principal claims.

## Bicep Skeleton

Review and customize:

```sh
deployments/azure-container-apps-indexer/infra/main.bicep
deployments/azure-container-apps-indexer/infra/main.parameters.example.json
```

Local render check:

```sh
az bicep build --file deployments/azure-container-apps-indexer/infra/main.bicep
```

Deployment preview:

```sh
az deployment group what-if \
  --resource-group <resource-group> \
  --template-file deployments/azure-container-apps-indexer/infra/main.bicep \
  --parameters @deployments/azure-container-apps-indexer/infra/main.parameters.example.json
```

Do not put secrets in `main.parameters.example.json`.

## Cairn Workspace Config Values

The workspace can contain non-secret values:

```yaml
workspace_id: cairn:workspace:pod-a

remote_sync:
  provider: azure_blob
  account: examplecairn
  container: pod-a
  prefix: ""

remote_index:
  url: https://<container-app-fqdn>
  audience: api://cairn-indexer
  tenant_id: <tenant-id>
```

Do not store:

- Storage account keys.
- Client secrets.
- PostgreSQL passwords.
- Bearer tokens.

## Smoke Tests

After deployment:

```sh
az login --tenant <tenant-id>
cairn validate
cairn sync status
cairn sync dry-run
cairn index status
cairn index refresh
cairn search "known term"
```

Expected behavior:

- Missing or invalid auth reports a remote auth warning, not local data loss.
- Remote sync refuses divergence instead of merging.
- Remote index outages degrade search while local search still works.
- Index refresh returns accepted or refreshed state with a next step.

## Follow-Up Implementation Work

- Add ACA built-in auth configuration to the Bicep module.
- Add indexer authorization middleware or shim for workspace/pod checks.
- Replace the local indexer prototype with a production CocoIndex pipeline image.
- Add a production runbook for refresh failures and stale index freshness.
- Add deployment-specific monitoring and alerts.
