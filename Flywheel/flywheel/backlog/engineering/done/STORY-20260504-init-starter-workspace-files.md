# Story: Init Starter Workspace Files

## Metadata
- `id`: STORY-20260504-init-starter-workspace-files
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-document-model-lifecycle, ADR-mcp-operation-surface]
- `success_metric`: `cairn init` creates the v1 starter workspace files expected by the north star.
- `release_scope`: required

## Problem Statement
- `cairn init` creates a config and folders, but the north star expects starter `.cairnignore`, schemas, onboarding, and minimal agent pointers.

## Scope
- In:
  - Create `.cairnignore` with sensible defaults.
  - Create starter schema files under `.cairn/schemas/`.
  - Create starter onboarding/setup documents.
  - Create terse `AGENTS.md` and `CLAUDE.md` when missing.
  - Keep init idempotent and non-destructive.
- Out:
  - Interactive prompts.
  - Team-specific setup content.
  - Secret-bearing profile setup.

## Assumptions
- Existing files should be preserved and reported as existing, not overwritten.

## Acceptance Criteria
1. Fresh `cairn init` creates expected starter files.
2. Re-running init is idempotent.
3. Existing `AGENTS.md` or `CLAUDE.md` is not overwritten.
4. CLI output reports completed work and next suggested step.

## Validation
- Required checks:
  - Workspace init tests.
  - CLI init tests.
  - Full `go test ./...`.

## Dependencies
- `STORY-20260503-workspace-init-config`
- `STORY-20260503-config-yaml-schema-validation`

## Risks
- Starter docs can become too verbose; keep them pointer-sized.

## Open Questions
- Resolved by existing implementation: `.cairn/schemas/core.yaml`, `.cairn/schemas/README.md`, `onboarding/team-context.md`, `onboarding/agent-setup.md`, and `onboarding/workspace-map.md`.

## Next Step
- Promote `STORY-20260504-aca-indexer-infra-module-skeleton`.

## Handoff Notes
- Engineering verified 2026-05-04.
- Existing `cairn init` already creates `.cairnignore`, starter schemas, onboarding docs, `AGENTS.md`, and `CLAUDE.md`.
- Existing implementation is idempotent and non-destructive for pre-existing files.
- Existing CLI output reports created/existing paths and a next validation step.
- QA completed 2026-05-04 with focused workspace/CLI init tests and full `go test ./...`.
