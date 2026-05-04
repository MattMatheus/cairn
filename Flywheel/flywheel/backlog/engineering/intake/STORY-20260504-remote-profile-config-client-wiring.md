# Story: Remote Profile Config Client Wiring

## Metadata
- `id`: STORY-20260504-remote-profile-config-client-wiring
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Cairn can construct remote sync and index clients from workspace config without storing secrets.
- `release_scope`: required

## Problem Statement
- Remote store and remote index clients exist, but CLI/MCP local operations do not yet load `.cairn/config.yaml` profile settings to configure them.

## Scope
- In:
  - Extend config loading for remote sync and remote index settings.
  - Build `remotestore.AzureBlobStore` from configured account/container/prefix.
  - Build `remoteindex.HTTPClient` from configured URL/audience fields.
  - Use Azure CLI token provider boundaries without storing secrets in workspace config.
  - Wire configured clients into CLI/MCP operation setup.
  - Preserve graceful degradation when remote config is absent or incomplete.
  - Add tests with fake token/client/store boundaries where practical.
- Out:
  - Live Azure calls.
  - Terraform/Bicep deployment.
  - Credential persistence.

## Assumptions
- Existing generated config may need additive fields; old config should remain loadable.

## Acceptance Criteria
1. Config can represent Azure Blob sync settings and remote index settings.
2. Missing config degrades to local-only behavior.
3. Configured sync operations receive a remote store.
4. Configured index operations receive a remote index client.
5. Workspace config does not require or store client secrets, account keys, or long-lived tokens.

## Validation
- Required checks:
  - Unit tests for config parsing and client construction.
  - CLI/MCP operation tests using fakes where practical.

## Dependencies
- `STORY-20260503-azure-blob-sync-adapter`
- `STORY-20260503-remote-index-search-integration`
- `STORY-20260503-config-yaml-schema-validation`

## Risks
- Keep profile config additive and backward-compatible with existing workspaces.

## Open Questions
- Whether indexer audience belongs in `.cairn/config.yaml` or an external profile file long term.

## Next Step
- Engineering should implement after gated remote MCP surface or in parallel if kept transport-neutral.
