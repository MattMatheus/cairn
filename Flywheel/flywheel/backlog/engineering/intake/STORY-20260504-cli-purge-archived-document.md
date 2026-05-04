# Story: CLI-Only Purge Archived Document

## Metadata
- `id`: STORY-20260504-cli-purge-archived-document
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-document-model-lifecycle, ADR-mcp-operation-surface]
- `success_metric`: Cairn supports explicit CLI-only hard deletion for archived documents without exposing purge through MCP.
- `release_scope`: required

## Problem Statement
- Archive exists, but the ADR also defines hard deletion/purge as a separate CLI-only operation requiring explicit confirmation.

## Scope
- In:
  - Implement document purge for archived documents only.
  - Require explicit CLI confirmation flag or typed confirmation.
  - Refuse purge for non-archived documents.
  - Preserve MCP absence of purge/delete tools.
  - Add tests for successful purge, refusal, and MCP non-exposure.
- Out:
  - Retention policy.
  - Soft-delete beyond archive.
  - MCP purge/delete.

## Assumptions
- V1 purge can remove the local file only; remote deletion follows normal sync push behavior.

## Acceptance Criteria
1. `cairn purge` exists and is CLI-only.
2. Purge requires explicit confirmation.
3. Purge refuses non-archived documents.
4. Successful purge removes the archived file and reports next steps.
5. MCP still exposes no purge/delete tool in any mode.

## Validation
- Required checks:
  - Document lifecycle tests.
  - CLI tests.
  - MCP registration tests.

## Dependencies
- `STORY-20260502-capture-promotion-archive`
- `STORY-20260503-mcp-mutating-tools-gated`

## Risks
- Avoid making hard delete too easy to trigger from agent workflows.

## Open Questions
- Confirmation UX: `--confirm-purge` versus typed document id.

## Next Step
- Engineering should implement after sync validation gate or independently.
