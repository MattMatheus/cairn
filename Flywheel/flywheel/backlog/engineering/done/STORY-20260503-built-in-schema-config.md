# Story: Built-In Schema And Type Config Foundation

## Metadata
- `id`: STORY-20260503-built-in-schema-config
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-document-model-lifecycle]
- `success_metric`: Cairn can load built-in/custom document type configuration and use it for destination folder conventions.
- `release_scope`: required

## Problem Statement
- The north star references built-in schemas and workspace config mapping document types to destination folders, but lifecycle code currently relies on fixed defaults.

## Scope
- In:
  - Define minimal config structure for built-in document types and destination folders.
  - Load config from `/.cairn/config.yaml` with defaults.
  - Use config in promotion target selection where practical.
  - Add tests for default and custom destination mapping.
- Out:
  - Full custom schema validation.
  - Interactive config editing.
  - Remote profile config.

## Assumptions
- YAML dependency choice should be small and isolated.

## Acceptance Criteria
1. Defaults cover built-in v1 document types.
2. Config load falls back safely when config is absent.
3. Promotion destination can use configured folder mappings.
4. Tests cover defaults and custom mappings.

## Validation
- Required checks:
  - Unit tests for config loading and destination mapping.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against document lifecycle ADR and north-star config notes.

## Dependencies
- `STORY-20260503-workspace-init-config`

## Risks
- Avoid building full schema validation too early.

## Open Questions
- Resolved for v1 foundation: use a tiny isolated parser for the simple generated `.cairn/config.yaml` shape; defer a YAML dependency until full custom schema validation needs it.

## Next Step
- Engineering should implement config loading and destination mapping.

## PM Handoff
- Promoted on 2026-05-03 as the last current intake story.
- Keep scope to built-in document types and destination folder mappings.
- Do not implement full custom schema validation in this slice.

## Engineering Handoff
- Added `document.Config`, `DefaultConfig`, and `LoadConfig`.
- Defaults cover built-in v1 document types and managed folders.
- Config loading falls back to defaults when `.cairn/config.yaml` is absent.
- Implemented an isolated minimal parser for the simple generated config shape; no YAML dependency added yet.
- Promotion now uses configured document type destination folders.
- Workspace validation uses configured managed folders when deciding whether markdown without frontmatter is Cairn-managed.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/document ./internal/workspace`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed defaults cover built-in v1 document types and managed folders.
- Confirmed config loading falls back safely when `.cairn/config.yaml` is absent.
- Confirmed promotion uses configured destination folder mappings.
- Confirmed validation treats configured managed folders, including nested folders, as managed markdown locations.
- Confirmed full custom schema validation remains out of scope.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Return to PM backlog planning. Current intake is empty, so the next cycle should create follow-up stories for sync pull/push orchestration, remote index integration, or packaging/deployment.
