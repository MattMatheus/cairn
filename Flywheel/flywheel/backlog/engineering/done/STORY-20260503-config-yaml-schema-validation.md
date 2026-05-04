# Story: Config YAML Schema Validation

## Metadata
- `id`: STORY-20260503-config-yaml-schema-validation
- `owner_role`: Software Architect
- `status`: done
- `source`: pm
- `decision_refs`: [ADR-document-model-lifecycle]
- `success_metric`: Cairn validates workspace config and custom schema files enough to catch invalid core fields and destination mappings.
- `release_scope`: optional

## Problem Statement
- Cairn can load the simple generated config shape, but it does not validate malformed config or custom schema YAML.

## Scope
- In:
  - Introduce a real YAML parser if justified.
  - Validate `.cairn/config.yaml` required/core fields.
  - Validate custom schema files preserve required Cairn core fields.
  - Surface config/schema findings through `validate_workspace`.
  - Add tests for malformed config and invalid custom schemas.
- Out:
  - Rich schema-driven document validation.
  - Interactive config editing.
  - Remote profile credential validation.

## Assumptions
- This should follow the simple config loader foundation, not replace it prematurely.

## Acceptance Criteria
1. Malformed config produces validation findings.
2. Unknown document type mappings warn but do not crash.
3. Custom schema missing core fields produces validation findings.
4. `validate_workspace` surfaces config/schema findings.

## Validation
- Required checks:
  - Unit tests for config/schema validation.
  - Repository formatting/lint checks if configured.

## Dependencies
- `STORY-20260503-built-in-schema-config`

## Risks
- Avoid turning this into a full custom schema engine.

## Open Questions
- YAML library choice for broader validation.

## Engineering Handoff
- Implemented 2026-05-03.
- Added `document.ValidateConfigFiles`.
- Added lightweight validation for `.cairn/config.yaml` required keys, scalar values, managed folders, and document type destinations.
- Unknown document type mappings warn without crashing.
- Added custom schema validation for `.cairn/schemas/*.yaml` to ensure required Cairn core fields remain present.
- Wired config/schema findings into `workspace.Validate` and `mcpops.ValidateWorkspace`.

## QA Handoff
- Accepted 2026-05-03.
- Verified malformed config produces validation errors.
- Verified unknown document type mappings produce warnings.
- Verified custom schemas missing Cairn core fields produce validation errors.
- Verified `validate_workspace` surfaces config/schema findings through the common envelope.
- Verification: `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Step
- Planning needed: current active backlog batch is complete.
