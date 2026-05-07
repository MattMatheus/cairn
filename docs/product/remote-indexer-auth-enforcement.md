# Remote Indexer Auth Enforcement Spike

## Recommendation

Historical/deferred production auth note: the earlier production direction used Azure Container Apps built-in authentication with Microsoft Entra ID as the v1 edge authentication mechanism, configured to require authenticated HTTPS traffic. The active local development path does not use ACA or Entra simulation; see `docs/adr/ADR-local-development-emulation.md`.

Inside the indexer, perform lightweight authorization from platform-provided authenticated identity data:

- Validate the request has passed the production edge authentication layer.
- Read forwarded principal/claim headers from that edge layer.
- Check tenant, subject/app identity, and allowed group or app-role claims against pod/workspace authorization config.
- Check the request workspace id matches the configured `CAIRN_INDEXER_WORKSPACE_ID`.

Add custom JWT validation middleware only if a real tenant test shows the production edge auth layer does not expose enough claims for pod/workspace authorization or does not support the non-browser bearer-token flow cleanly.

## Why This Path

- It keeps authentication at the platform edge and reduces bespoke security code in the indexer.
- It preserves the north-star constraint that Cairn workspace config must not store secrets.
- It leaves authorization in the indexer, where workspace and pod-specific checks belong.
- It keeps Entra validation outside the core local development harness.

## Required Checks

Token or principal data must support:

- `aud`: configured indexer audience or app id URI.
- `iss`: expected Microsoft Entra issuer for the tenant.
- `tid`: expected tenant id.
- `sub`, `oid`, or app id: stable caller identity.
- Group or app-role claim when tenant policy requires scoped workspace access.
- Workspace/pod claim or server-side allow-list entry tying the caller to `CAIRN_INDEXER_WORKSPACE_ID`.

## Failure Mapping

| Condition | Indexer/edge response | Cairn behavior |
| --- | --- | --- |
| Missing token | `401` | Report remote auth unavailable and suggest Azure login/profile check |
| Invalid token | `401` | Report remote auth unavailable and suggest token refresh |
| Valid identity, unauthorized workspace | `403` | Report remote service/auth warning without local failure |
| Production auth service unavailable | `503` or request failure | Degrade remote index mode and suggest retry/status |
| Claims missing for authorization | `403` | Report workspace unauthorized and point to operator config |

## Implementation Follow-Up

Future production work should create a narrow story for the selected hosting layer's auth configuration and an indexer authorization shim that consumes authenticated principal headers. That story should include an integration note for validating Azure CLI acquired tokens against the deployed endpoint.

## Sources

- Historical references: Azure Container Apps authentication and authorization; managed identities in Azure Container Apps, accessed 2026-05-04.
