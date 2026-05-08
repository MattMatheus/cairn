# Story: Interactive Capture Flow

## Metadata
- `id`: STORY-20260508-interactive-capture-flow
- `owner_role`: Product Manager
- `status`: done
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

## PM Refinement

- First implementation should add a short human-friendly CLI path, preferably `cairn note`, for common manual capture.
- `cairn note` should default actor from `CAIRN_ACTOR`, `USER`, or `USERNAME` when `--actor` is not supplied.
- Type support should include `note`, `investigation`, `handoff`, `decision`, and `runbook`.
- Existing non-interactive `cairn capture` behavior must remain compatible.

## Next Step

QA should validate the short `note` path, prompt-driven `capture --interactive`, type support, and existing capture compatibility.

## Engineering Handoff

### What Changed

- Added `cairn note` as a short human-friendly capture path.
- `cairn note` defaults actor from `CAIRN_ACTOR`, `USER`, or `USERNAME`.
- `cairn note` supports `note`, `investigation`, `handoff`, `decision`, and `runbook` types with starter markdown templates when no body is supplied.
- Added `cairn capture --interactive` to prompt for missing actor, title, type, and multiline body.
- Preserved existing non-interactive `cairn capture` behavior.
- Capture next steps now include validate, promote, and sync.

### Validation Evidence

- `go test ./...` passed.
- Manual smoke: `CAIRN_ACTOR='Pilot Dev' cairn note --title "Restart Worker" --type runbook` created `agents/pilot-dev/restart-worker.md` with runbook template.
- Manual smoke: prompt-driven `cairn capture --interactive` created `agents/pilot/interactive-handoff.md`.

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, or production actions were taken.
- No human approval was required beyond the user's request to continue.

### Risks And Assumptions

- `cairn note` intentionally stays small and template-driven instead of launching `$EDITOR` in this first slice.
- Actor defaulting slugifies environment values so path segments remain safe.
- Interactive prompts are line-oriented to keep them testable and shell-friendly.

### QA Focus Areas

- Verify `cairn note` is short enough for non-AI users and creates valid managed docs.
- Verify common types produce the expected `type` frontmatter.
- Verify `capture --interactive` can be driven from stdin and writes body content.
- Confirm existing `capture` tests and behavior still pass.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `go test ./...` passed.
- Manual QA smoke: `CAIRN_ACTOR='QA Dev' cairn note --title "QA Runbook" --type runbook` created a valid runbook capture and `cairn validate` passed.
- Manual QA smoke: stdin-driven `cairn capture --interactive` created a valid decision capture and `cairn validate` passed.

### Defects

None filed.

### Evidence Quality

Sufficient. Automated coverage exercises short note creation, all supported common capture types, interactive prompting, and existing capture compatibility.

### State Transition

Moved to `done`.
