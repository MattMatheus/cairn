# Story: Sync Validation Gate

## Metadata
- `id`: STORY-20260504-sync-validation-gate
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-document-model-lifecycle, ADR-sync-conflict-behavior]
- `success_metric`: Sync refuses invalid managed markdown before mutating local or remote state.
- `release_scope`: required

## Problem Statement
- The sync ADR says document sync requires valid core frontmatter for managed markdown unless ignored, but current pull/push logic focuses on manifest divergence and object IO.

## Scope
- In:
  - Add a reusable durable-boundary validation gate for sync operations.
  - Refuse `sync_push` when local managed markdown has blocking validation findings.
  - Refuse `sync_pull` when the accepted remote manifest/objects would introduce invalid managed markdown where practical.
  - Include validation findings in warnings or error details without mutating state.
  - Add tests proving state and files are unchanged on validation refusal.
- Out:
  - Rich custom schema enforcement beyond current config/schema guardrails.
  - Automatic repair during sync.

## Assumptions
- Push validation can validate local files directly.
- Pull validation may need to validate fetched remote markdown before writing it locally.

## Acceptance Criteria
1. Invalid local managed markdown blocks sync push before remote writes.
2. Invalid remote managed markdown blocks sync pull before local writes/state updates.
3. Refusal response includes actionable validation findings.
4. Ignored files remain outside the sync validation gate.
5. Tests cover push refusal, pull refusal, and non-mutation guarantees.

## Validation
- Required checks:
  - Unit tests for sync validation refusal paths.
  - Repository formatting/lint checks if configured.

## Dependencies
- `STORY-20260503-sync-pull-apply`
- `STORY-20260503-sync-push-apply`
- `STORY-20260503-config-yaml-schema-validation`

## Risks
- Avoid fetching remote objects twice unless needed for validation and apply.

## Open Questions
- Whether pull validation should stage fetched content in memory or temporary files.

## Engineering Handoff
- Implemented 2026-05-04.
- Added reusable sync validation findings and `ValidationError`.
- `sync_push` validates generated local managed markdown manifest before any remote writes.
- `sync_pull` fetches and validates remote markdown in memory before any local file writes or sync state updates.
- Ignored files remain outside the sync validation gate through existing manifest generation rules.

## QA Handoff
- Accepted 2026-05-04.
- Verified invalid local managed markdown blocks push before remote writes.
- Verified invalid remote managed markdown blocks pull before local writes/state updates.
- Verified validation refusals return actionable path/message details.
- Verified ignored invalid markdown does not block valid sync changes.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Promote `STORY-20260504-mcp-remote-mutating-tools-gated`.
