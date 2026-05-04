# Story: Workspace Init And Config Foundation

## Metadata
- `id`: STORY-20260503-workspace-init-config
- `owner_role`: Software Architect
- `status`: done
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
- Resolved for v1: config starts with schema version, workspace id, managed folders, document type destinations, local profile, and pod-remote placeholders with no secrets.

## Next Step
- Engineering should implement idempotent non-interactive workspace init.

## PM Handoff
- Promoted on 2026-05-03 after workspace validation completed.
- Keep init non-interactive and conservative; no credential setup or rich onboarding generation.
- Make created files useful as pointers while avoiding overwrites of existing human-authored content.

## Engineering Handoff
- Implemented `internal/workspace.Init` with idempotent creation of the north-star folder layout.
- Added minimal generated `.cairn/config.yaml`, `.cairnignore`, schema README, onboarding files, `AGENTS.md`, and `CLAUDE.md`.
- Existing files are never overwritten; conflicting directory/file paths return an error.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/workspace`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed init creates the north-star layout, control files, starter onboarding, agent pointers, and schema starter content.
- Confirmed existing files are not overwritten and repeated init is idempotent.
- QA polish applied before acceptance: added `.cairn/schemas/core.yaml` to satisfy starter schema intent and replaced custom key sorting with `sort.Strings`.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote the local CLI command surface story so init and validation can be invoked from user-facing commands.
