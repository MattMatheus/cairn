# Story: Built-In Schema And Type Config Foundation

## Metadata
- `id`: STORY-20260503-built-in-schema-config
- `owner_role`: Software Architect
- `status`: intake
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
- YAML package choice.

## Next Step
- PM should refine after init/config foundation.
