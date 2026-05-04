# Story: Init Starter Workspace Files

## Metadata
- `id`: STORY-20260504-init-starter-workspace-files
- `owner_role`: Software Architect
- `status`: intake
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
- Exact starter onboarding path and schema filenames.

## Next Step
- PM should refine the starter file list before engineering.
