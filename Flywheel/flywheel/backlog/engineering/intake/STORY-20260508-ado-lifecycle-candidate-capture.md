# Story: ADO Lifecycle Candidate Capture

## Metadata
- `id`: STORY-20260508-ado-lifecycle-candidate-capture
- `owner_role`: Product Manager
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-20260502-document-model-lifecycle]
- `success_metric`: A configured ADO lifecycle event can create a validated working/proposed Cairn candidate note without promoting it to canonical automatically.
- `release_scope`: deferred

## Problem Statement

ADO is the expected enterprise workflow surface. Cairn should eventually turn meaningful work item, PR, incident, or release transitions into candidate knowledge without making developers duplicate lifecycle context by hand.

## Scope
- In: Define and implement an initial ADO lifecycle hook path that captures candidate Cairn knowledge.
- In: Include ADO metadata such as work item, PR, repo, branch, transition, and actor when available.
- In: Create working or proposed documents only.
- Out: Automatic canonical promotion, AI review, PR blocking as the primary safety gate, or central telemetry.

## Assumptions

- Candidate capture should use Cairn CLI or MCP mutation gates so validation remains centralized.
- The first hook can be narrow and configurable rather than trying to cover every ADO state transition.

## Acceptance Criteria
1. A configured ADO lifecycle event can create a candidate Cairn note with durable frontmatter and source metadata.
2. The generated candidate clearly identifies its ADO origin and recommended next action.
3. The flow never creates canonical knowledge without an explicit Cairn promotion.

## Validation
- Required checks: `go test ./...`
- Additional checks: local or mocked ADO payload fixture for the selected initial transition.

## Dependencies

- Interactive capture and repo attachment stories are expected to reduce ambiguity before ADO hooks are implemented.

## Risks

- Hooks may generate low-value noise. Start with explicit state transitions and make creation opt-in/configurable.
- ADO auth and payload variance can expand scope. Keep v1 to a minimal integration shape.

## Open Questions

- Which lifecycle transition should be first: PR completed, work item closed, release succeeded, or incident closed?

## Next Step

Keep in intake until local CLI polish and repo discovery are clearer.
