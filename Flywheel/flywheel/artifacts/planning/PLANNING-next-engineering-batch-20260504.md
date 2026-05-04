# Planning: Next Engineering Batch

## Objective

Create the next engineering intake batch after completing the first ADR-driven implementation run.

## Batch Theme

Finish the operational surface around already-built primitives before expanding product scope:

- make local index refresh directly usable
- enforce sync and lifecycle safety gates
- expose remote mutations only behind explicit MCP opt-in
- wire remote profile configuration without storing secrets
- complete CLI-only purge as the hard-delete boundary

## Proposed Order

1. `STORY-20260504-local-index-refresh-command`
2. `STORY-20260504-sync-validation-gate`
3. `STORY-20260504-mcp-remote-mutating-tools-gated`
4. `STORY-20260504-remote-profile-config-client-wiring`
5. `STORY-20260504-cli-purge-archived-document`
6. `STORY-20260504-lifecycle-transition-enforcement`

## Rationale

- Local index refresh improves immediate daily usability and has a small blast radius.
- Sync validation should land before sync mutations are exposed broadly to agents.
- Remote MCP mutation gating can reuse completed sync/index adapters once validation is safer.
- Remote profile config makes the remote surface practical without secrets in workspace files.
- Purge and transition enforcement close lifecycle gaps from the document ADR.

## Deferred

- Production ACA automation remains deferred until the documented plan becomes a priority.
- Rich custom schema validation remains out of scope.
- Retention policy and privacy filtering remain v2/later safeguards.

## Validation

- Stories were checked against accepted ADRs:
  - `ADR-document-model-lifecycle`
  - `ADR-mcp-operation-surface`
  - `ADR-sync-conflict-behavior`
  - `ADR-indexing-query-boundary`

## Next Step

Promote and implement `STORY-20260504-local-index-refresh-command`.
