# Story: Lifecycle Transition Enforcement

## Metadata
- `id`: STORY-20260504-lifecycle-transition-enforcement
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-document-model-lifecycle]
- `success_metric`: Promotion/archive operations enforce the accepted v1 status transition model.
- `release_scope`: required

## Problem Statement
- The ADR defines allowed document lifecycle transitions, but lifecycle operations need explicit tests and enforcement for invalid jumps.

## Scope
- In:
  - Centralize allowed status transitions.
  - Enforce promotion transitions such as draft/proposed/canonical boundaries.
  - Keep archive as allowed from any state.
  - Preserve proposed promotion repair behavior where allowed.
  - Add validation tests for allowed and refused transitions.
- Out:
  - Workflow approvals.
  - Multi-user review state.
  - Retention policy.

## Assumptions
- Existing capture default status remains `working`.
- Canonical promotion should not silently skip invalid intermediate states unless explicitly allowed by ADR.

## Acceptance Criteria
1. Allowed transitions match the document lifecycle ADR.
2. Invalid promotion transitions refuse without moving or rewriting files.
3. Archive remains allowed from any managed document state.
4. Tests cover working/proposed/canonical/archive paths.

## Validation
- Required checks:
  - Document lifecycle unit tests.
  - CLI/MCP adapter tests if behavior changes surface there.

## Dependencies
- `STORY-20260502-capture-promotion-archive`

## Risks
- Tightening transitions may break tests that relied on permissive canonical promotion.

## Open Questions
- Whether direct `working -> canonical` should remain supported as a convenience or require `working -> proposed -> canonical`.

## Next Step
- PM/engineering should answer the open transition question before implementation.
