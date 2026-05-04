# Story: Remote Indexer Auth Enforcement Spike

## Metadata
- `id`: STORY-20260504-remote-indexer-auth-enforcement-spike
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: Cairn has a concrete recommendation for ACA built-in auth versus indexer middleware token validation.
- `release_scope`: optional

## Problem Statement
- The indexing ADR and ACA plan leave exact auth enforcement as a follow-up detail, while Cairn already assumes Azure CLI bearer tokens for remote index calls.

## Scope
- In:
  - Compare ACA built-in auth with in-service JWT validation middleware.
  - Identify required token claims and workspace/pod authorization checks.
  - Define failure mapping for 401, 403, unavailable, and stale index behavior.
  - Produce a short decision note or ADR addendum candidate.
- Out:
  - Live Entra app registration.
  - Full middleware implementation.
  - Secrets or tenant-specific values.

## Assumptions
- The preferred path should keep Cairn config secret-free and use Azure CLI identity.

## Acceptance Criteria
1. Recommendation is documented with tradeoffs.
2. Required token audience/issuer/tenant/workspace checks are listed.
3. Failure responses map to Cairn warnings and next steps.
4. Follow-up implementation story is identified if needed.

## Validation
- Required checks:
  - Manual review against north star auth constraints and ACA deployment plan.

## Dependencies
- `STORY-20260504-remote-profile-config-client-wiring`
- `STORY-20260503-aca-indexer-deployment-plan`

## Risks
- Overfitting to one enterprise tenant setup too early.

## Open Questions
- Whether ACA built-in auth exposes enough claims to avoid custom middleware in v1.

## Next Step
- PM should decide whether this needs an ADR addendum or can remain an implementation spike.
