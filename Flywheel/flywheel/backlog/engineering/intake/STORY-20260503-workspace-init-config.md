# Story: Workspace Init And Config Foundation

## Metadata
- `id`: STORY-20260503-workspace-init-config
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-document-model-lifecycle, ADR-sync-conflict-behavior, ADR-indexing-query-boundary]
- `success_metric`: Cairn can initialize the standard workspace layout and local config files non-interactively.
- `release_scope`: required

## Problem Statement
- Product behavior assumes a standard workspace layout, `.cairn` control directory, `.cairnignore`, config, and starter onboarding files, but there is no init operation yet.

## Scope
- In:
  - Implement reusable non-interactive init operation.
  - Create standard top-level folders.
  - Create `.cairn/config.yaml`, `.cairnignore`, starter schemas directory, and starter onboarding files.
  - Create minimal root `AGENTS.md` and `CLAUDE.md` if absent.
  - Make init idempotent and non-destructive.
  - Add tests for new and partially existing workspaces.
- Out:
  - Interactive CLI prompts.
  - Remote profile credential setup.
  - Rich generated onboarding content.

## Assumptions
- Config schema can start minimal and grow with sync/profile stories.

## Acceptance Criteria
1. Init creates the north-star folder layout.
2. Init creates minimal `.cairn/config.yaml` and `.cairnignore`.
3. Init creates starter onboarding and agent pointer files without overwriting existing content.
4. Init is idempotent.
5. Tests cover fresh and partially initialized workspaces.

## Validation
- Required checks:
  - Unit/integration tests for init behavior.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against `docs/product/north-star.md`.

## Dependencies
- None beyond current document primitives.

## Risks
- Avoid overfitting config before sync/profile behavior lands.

## Open Questions
- Exact minimal config fields.

## Next Step
- PM should refine after current MCP adapter slices.
