# Story: Validate Workspace Operation

## Metadata
- `id`: STORY-20260503-validate-workspace-operation
- `owner_role`: Software Architect
- `status`: intake
- `source`: pm
- `decision_refs`: [ADR-document-model-lifecycle, ADR-mcp-operation-surface, ADR-sync-conflict-behavior, ADR-indexing-query-boundary]
- `success_metric`: Cairn can validate managed markdown and local sync/index metadata health through a reusable operation.
- `release_scope`: required

## Problem Statement
- Cairn has document validation primitives, sync state, and index artifacts, but no workspace-level operation that agents or future CLI/MCP surfaces can call.

## Scope
- In:
  - Implement reusable workspace validation over managed markdown.
  - Report document frontmatter findings using `mcpschema.ValidateWorkspaceData`.
  - Include lightweight checks for local sync state and local index availability.
  - Respect `.cairnignore` where available.
  - Add tests for missing/invalid frontmatter and metadata health warnings.
- Out:
  - Full schema YAML validation.
  - Remote sync/index health checks.
  - CLI/MCP transport wiring.

## Assumptions
- Discovery remains permissive but durable operations stay strict.

## Acceptance Criteria
1. Managed markdown validation returns findings with severity, code, message, path, and document id where available.
2. Missing/invalid frontmatter produces warnings or errors consistent with validation mode.
3. Local sync/index health warnings are surfaced.
4. Ignored paths are skipped.
5. Tests cover representative healthy and unhealthy workspaces.

## Validation
- Required checks:
  - Unit/integration tests for workspace validation.
  - Repository formatting/lint checks if configured.
- Additional checks:
  - Manual review against lifecycle, sync, index, and MCP ADRs.

## Dependencies
- `STORY-20260502-document-frontmatter-validation`
- `STORY-20260502-sync-manifest-state`
- `STORY-20260502-local-metadata-index`

## Risks
- Keep validation bounded; custom schema YAML can follow later.

## Open Questions
- Should this live in `internal/document` or a new workspace package?

## Next Step
- PM should refine package ownership before promotion.
