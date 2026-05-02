# HUMANS

Operator guide for the Flywheel workflow harness.

## Purpose
Flywheel provides a staged workflow for planning, architecture, implementation, QA, PM refinement, and cycle closure.

The harness does not require a product-specific repo layout. Operators should use the locations defined in `flywheel.yaml`.

## Operator Entry
1. Read `flywheel.yaml`.
2. Confirm the current branch matches `workflow.required_branch`.
3. Read `flywheel/DEVELOPMENT_CYCLE.md`.
4. Open the configured engineering active queue.
5. Launch the required stage through `flywheel/tools/launch_stage.sh`.

## Stage Order
- Planning: shape new work and create intake artifacts.
- Architect: process architecture/design decision work.
- Engineering: implement the top active engineering story.
- QA: validate the engineering story and decide `done` or return to `active`.
- PM: refine intake and maintain queue order.
- Cycle: alternate engineering and QA until the active queue is drained.

## Core Rules
- Do not invent work when the configured active queue is empty.
- Use explicit backlog state transitions.
- Do not commit during intermediate stage transitions.
- Close each completed cycle with an observer report.
- Use one commit per completed cycle.
- Treat artifact readiness as explicit, not implied.
- Record validation evidence, open risks, and next-state recommendation in handoffs.
- Use the smallest useful action model:
  - `read`
  - `local write`
  - `risky write`
  - `sensitive or production`
- Require explicit human approval for `risky write` and `sensitive or production` actions.
- Record approval outcome when approval-governed work occurs.
- Keep prompts, process docs, and tool behavior synchronized when the workflow changes.
- Treat the artifact tool as an optional integration, not required Flywheel core behavior, unless the repo explicitly enables it in config.

## Config-Owned Surfaces
These locations are owned by `flywheel.yaml`:
- prompt directory
- role directory
- process directory
- template directory
- planning artifact directory
- observer artifact directory
- engineering lanes
- architecture lanes

## Minimal Operator Checklist
1. Select the stage.
2. Read the stage prompt from the configured prompt directory.
3. Work only in the configured backlog lanes and artifact directories.
4. Produce the required handoff for the stage.
5. Run `flywheel/tools/run_observer_cycle.sh` at cycle closure.
6. Commit using `workflow.cycle_commit_format`.

If `integrations.artifact_workflow.enabled` is `true`, use the artifact-tool commands surfaced by `launch_stage.sh` and `run_observer_cycle.sh` when they help with artifact selection or durable handoff records.

If you need to automate around those hints, `flywheel/tools/artifact_workflow.sh --format json` provides the same guidance in machine-readable form.

When agents are running a stage, prefer telling them to consult `flywheel/tools/artifact_workflow.sh <stage> --format json` at stage entry and exit instead of relying on prose-only reminders.

If the agent only needs one phase, prefer `flywheel/tools/artifact_workflow_commands.sh --stage <stage> --phase <entry|exit>` to avoid extra parsing.
