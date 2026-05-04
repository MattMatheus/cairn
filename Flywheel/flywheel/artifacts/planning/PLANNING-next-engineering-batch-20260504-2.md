# Planning: Next Engineering Batch

## Context
- The previous batch completed purge and lifecycle transition enforcement.
- The active/ready lanes are empty.
- The prior README candidates, ACA deployment plan and config YAML schema validation, are already done.

## Proposed Order
1. `.cairnignore` support for discovery, validation, indexing, and sync manifests.
2. Sync deletion and purge propagation semantics.
3. Non-interactive `cairn init` starter workspace outputs.
4. ACA indexer infrastructure module skeleton.
5. Remote indexer auth enforcement design spike.

## Rationale
- Ignore support is foundational for sync and validation correctness.
- Deletion semantics should follow after purge exists so hard delete stays CLI-only while sync can publish manifest changes.
- Init ergonomics reduce setup friction once config/schema validation exists.
- ACA automation can begin from the documented deployment plan without requiring a live deployment.
- Auth enforcement remains the riskiest remote-indexer detail and should be isolated from core local workflows.

## Next Suggested Step
- Promote `STORY-20260504-cairnignore-workspace-filtering`.
