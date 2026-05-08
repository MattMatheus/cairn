# Story: ADO Lifecycle Candidate Capture

## Metadata
- `id`: STORY-20260508-ado-lifecycle-candidate-capture
- `owner_role`: Product Manager
- `status`: done
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

Engineering should implement a narrow PR-completed candidate capture path that reads an ADO payload and creates only working or proposed Cairn knowledge.

## PM Refinement

- First lifecycle transition: ADO PR completed.
- Command shape: `cairn ado capture --event pr-completed --payload-file FILE`.
- Default actor should be `ado`.
- Default document type should be `handoff`.
- Default status should be `working`; optional `--status proposed` may promote through Cairn lifecycle.
- `canonical` must be rejected.
- Payload parsing should be tolerant and fixture-driven, not dependent on live ADO auth.

## Engineering Handoff

### What Changed

- Added `cairn ado capture --event pr-completed --payload-file FILE`.
- Added `--actor`, `--type`, and `--status working|proposed` options.
- Default actor is `ado`; default type is `handoff`; default status is `working`.
- Added tolerant fixture-driven ADO PR payload extraction for PR id, title, description, repository, branches, actor, and URL.
- Generated body clearly marks ADO origin and recommended next action.
- Canonical status is rejected; proposed status goes through Cairn capture plus promote lifecycle.

### Validation Evidence

- `go test ./...` passed.
- Manual mocked ADO PR payload smoke created a working candidate and `cairn validate` passed.
- Manual mocked ADO PR payload smoke created a proposed candidate and `cairn validate` passed.
- Manual canonical-status smoke failed with the expected rejection.

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, production, or live ADO actions were taken.
- No human approval was required beyond the user's request to continue.

### Risks And Assumptions

- First supported event is intentionally narrow: `pr-completed`.
- Payload parsing is tolerant and fixture-based; live ADO payload variations may need future expansion.
- Candidate body carries ADO source metadata; frontmatter `source` remains Cairn's lifecycle source.

### QA Focus Areas

- Verify a mocked PR-completed payload creates a valid working candidate.
- Verify `--status proposed` never skips Cairn lifecycle promotion.
- Verify `--status canonical` is rejected.
- Verify generated content clearly identifies the ADO origin and candidate review next action.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `go test ./...` passed.
- Mocked ADO PR-completed payload created a valid working candidate and `cairn validate` passed.
- Mocked ADO PR-completed payload created a proposed candidate through Cairn promotion and `cairn validate` passed.
- `--status canonical` failed as expected with `ado capture supports only working or proposed status`.

### Defects

None filed.

### Evidence Quality

Sufficient. Tests cover payload parsing, unsupported event rejection, working capture, proposed promotion, and canonical rejection.

### State Transition

Moved to `done`.
