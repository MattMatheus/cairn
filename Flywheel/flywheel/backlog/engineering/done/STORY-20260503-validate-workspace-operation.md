# Story: Validate Workspace Operation

## Metadata
- `id`: STORY-20260503-validate-workspace-operation
- `owner_role`: Software Architect
- `status`: done
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
- Resolved for v1: reusable operation lives in a new `internal/workspace` package with thin callers layered above it.

## Next Step
- Engineering should implement reusable workspace validation and run the story through QA.

## PM Handoff
- Promoted on 2026-05-03 after package ownership refinement.
- Keep schema validation bounded to existing document frontmatter validation.
- Use `mcpschema.ValidateWorkspaceData` as the operation response shape so MCP/CLI surfaces can reuse it later.

## Engineering Handoff
- Implemented `internal/workspace.Validate` as the reusable workspace validation operation.
- Added `.cairnignore`-aware markdown walking, document frontmatter findings, and local sync/index metadata health findings.
- Added `Local.ValidateWorkspace` as a thin MCP operation adapter with next-step guidance.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./internal/workspace ./internal/mcpops`
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Handoff
- Accepted on 2026-05-03.
- Confirmed acceptance criteria are covered by workspace validation tests:
  - document finding shape and severity by validation mode
  - local sync/index metadata health warnings
  - `.cairnignore` skipped paths
  - healthy workspace with valid metadata artifacts
- QA fix applied before acceptance: explicit requested paths that traverse outside the workspace are skipped.
- Verification:
  - `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## Next Suggested Step
- Promote the workspace init/config foundation story so validation can distinguish missing workspace setup from degraded workspace metadata.
