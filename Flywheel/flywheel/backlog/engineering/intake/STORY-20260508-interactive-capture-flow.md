# Story: Interactive Capture Flow

## Metadata
- `id`: STORY-20260508-interactive-capture-flow
- `owner_role`: Product Manager
- `status`: intake
- `source`: planning
- `decision_refs`: [ADR-20260502-document-model-lifecycle]
- `success_metric`: A non-AI developer can create a valid working/proposed Cairn note without writing frontmatter manually or remembering the full capture command.
- `release_scope`: required

## Problem Statement

Some pilot developers will not use AI assistants. Cairn capture should therefore feel approachable from a terminal or editor without requiring long flags, frontmatter knowledge, or manual folder choices.

## Scope
- In: Add an interactive or guided capture path for common document types.
- In: Support launching `$EDITOR` with a useful template or collecting title/type/body from prompts.
- In: Preserve existing non-interactive `cairn capture` behavior.
- Out: Automatic canonical promotion, AI-generated summaries, ADO integration, or VS Code extension work.

## Assumptions

- Existing lifecycle operations can create valid frontmatter once the title, actor, type, and body are known.
- The first version can focus on terminal/editor use before plugin surfaces.

## Acceptance Criteria
1. A developer can run a short command and produce a valid captured document under the expected actor/workspace path.
2. The flow supports at least note, investigation, handoff, decision, and runbook-oriented capture starts.
3. Output includes the created path and clear next actions for validate, promote, and sync.

## Validation
- Required checks: `go test ./...`
- Additional checks: manual CLI run using stdin/editor-safe test mode if interactive prompts are introduced.

## Dependencies

- Existing document capture and promotion lifecycle.

## Risks

- Interactive behavior can make automated tests brittle. Keep core prompt logic separable from terminal IO.
- Too many templates could add confusion. Start with a small, obvious set.

## Open Questions

- Should the short command be `cairn note`, `cairn capture --interactive`, or both?
- What default actor should be used when no actor is supplied?

## Next Step

PM should refine the command shape after `doctor --full` is queued.
