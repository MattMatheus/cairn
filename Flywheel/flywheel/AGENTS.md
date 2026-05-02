# AGENTS

Agent operating guide for the Flywheel workflow harness.

## Mission
Execute work through the configured Flywheel stages without relying on product-specific assumptions.

## First 5 Minutes
1. Read `flywheel.yaml`.
2. Read `flywheel/HUMANS.md`.
3. Read `flywheel/DEVELOPMENT_CYCLE.md`.
4. Read the stage prompt from `paths.prompts`.
5. Read the relevant role contract from `paths.roles` when role selection is enabled.
6. If `integrations.artifact_workflow.enabled` is `true`, read `flywheel/tools/artifact_workflow.sh <stage> --format json` for stage-specific artifact guidance.

## Stage Prompts
Flywheel expects stage prompts for:
- planning
- architect
- engineering
- qa
- pm
- cycle

These should be resolved from `paths.prompts`.

## Mandatory Rules
- Respect `workflow.required_branch`.
- Use only the configured backlog and artifact paths.
- Respect explicit backlog state transitions.
- Do not fabricate work when the active lane is empty.
- Do not commit during intermediate stage transitions.
- Use one commit per completed cycle.
- Generate an observer artifact through `flywheel/tools/run_observer_cycle.sh` at cycle closure.
- Treat QA as a gate, not a suggestion.
- Treat artifact readiness as explicit, not implied.
- Record evidence, risks, and next-state recommendation in stage handoffs.
- Treat the artifact tool as an optional integration and use it only when the repo config enables it.
- When the artifact workflow integration is enabled, treat `flywheel/tools/artifact_workflow.sh --format json` as the canonical machine-readable source for stage entry and exit artifact guidance.
- Interpret the JSON output consistently:
  - `entry` commands help you discover or read the artifact inputs for the current stage.
  - `exit` commands help you write durable handoff or cycle-closure records after the stage work is complete.
- If you only need one phase, prefer `flywheel/tools/artifact_workflow_commands.sh --stage <stage> --phase <entry|exit>` to avoid reparsing the full JSON payload.
- Use the smallest useful action model:
  - `read`
  - `local write`
  - `risky write`
  - `sensitive or production`
- Require explicit human approval for `risky write` and `sensitive or production` actions.
- Record approval outcome when approval-governed work occurs.

## Required Sync
When workflow behavior changes, update together:
- `flywheel/HUMANS.md`
- `flywheel/AGENTS.md`
- `flywheel/DEVELOPMENT_CYCLE.md`
- the affected prompts
- the affected process docs
- any affected tool behavior

## Scope
Flywheel defines workflow behavior, not product strategy.

If a host repository needs product planning, release management, metrics, or additional governance, those should be layered on top of Flywheel rather than embedded into the core harness.
