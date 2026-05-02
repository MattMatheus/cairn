# flywheel

Flywheel is a configurable workflow harness for human+agent delivery.

It is optimized for explicit queue movement, reviewable markdown artifacts, and cycle closure with durable observer records.

It preserves a staged operating model:
- planning
- architect
- engineering
- qa
- pm
- cycle

The harness is intended to run either:
- in a new repository with Flywheel-owned directories
- inside an existing repository by mapping Flywheel paths in `flywheel.yaml`

## Start
1. Review `flywheel.yaml`.
2. Read `flywheel/HUMANS.md`.
3. Read `flywheel/DEVELOPMENT_CYCLE.md`.
4. Read the stage prompts, process docs, and templates.
5. Populate the configured prompt, role, process, template, and backlog paths.
6. Run the harness tools from `flywheel/tools/` against the configured paths.

## Typical Use

### Add Flywheel To An Existing Repo
1. Copy `flywheel.yaml` to the repo root.
2. Copy the `flywheel/` directory into the repo.
3. Keep the default local paths if you want a self-contained workflow.
4. Edit `flywheel.yaml` only if you want Flywheel to point at existing repo directories.
5. Run `./flywheel/tools/run_doc_tests.sh`.

### Example Remap For An Existing Repo
If an existing project already has its own work directories, keep the Flywheel system under `flywheel/` and remap only the workflow-owned state:

```yaml
paths:
  prompts: "flywheel/prompts"
  roles: "flywheel/roles"
  process: "flywheel/process"
  templates: "flywheel/templates"
  artifacts:
    planning: "ops/planning"
    observer: "ops/observer"
    release: "ops/release"
  engineering:
    intake: "work/engineering/intake"
    ready: "work/engineering/ready"
    active: "work/engineering/active"
    qa: "work/engineering/qa"
    done: "work/engineering/done"
    blocked: "work/engineering/blocked"
    archive: "work/engineering/archive"
  architecture:
    intake: "work/architecture/intake"
    ready: "work/architecture/ready"
    active: "work/architecture/active"
    qa: "work/architecture/qa"
    done: "work/architecture/done"
    blocked: "work/architecture/blocked"
    archive: "work/architecture/archive"
```

This keeps the harness system contained in `flywheel/` while allowing backlog state and generated artifacts to live in project-native locations.

### Use Flywheel As A Self-Contained Local Workflow
- backlog state lives under `flywheel/backlog/`
- planning notes live under `flywheel/artifacts/planning/`
- observer reports live under `flywheel/artifacts/observer/`
- prompts, roles, process docs, and templates stay under `flywheel/`
- tools run from `flywheel/tools/`

## Operating Expectations
- artifact readiness is explicit, not implied
- handoffs record evidence, risks, and next-state recommendation
- observer reports act as compact execution traces, not just end-of-cycle notes
- risky or sensitive actions require explicit approval and recorded outcome

### Launch A Stage
- `./flywheel/tools/launch_stage.sh planning`
- `./flywheel/tools/launch_stage.sh architect`
- `./flywheel/tools/launch_stage.sh engineering`
- `./flywheel/tools/launch_stage.sh qa`
- `./flywheel/tools/launch_stage.sh pm`
- `./flywheel/tools/launch_stage.sh cycle`

### Close A Cycle
- `./flywheel/tools/run_observer_cycle.sh --cycle-id <cycle-id>`

### Optional Artifact Workflow
Flywheel can surface artifact-tool commands without making that tool part of the harness contract.

Enable `integrations.artifact_workflow.enabled` in `flywheel.yaml` and point `integrations.artifact_workflow.command` at a wrapper such as `/Users/foundry/AgenticDevelopment/Tools/artifacts/flywheel-artifacts`.

When enabled:
- `launch_stage.sh` prints stage-specific artifact selection and manifest commands
- `run_observer_cycle.sh` prints a cycle-closure manifest command after writing the observer report
- `artifact_workflow.sh --format json` returns the same hints in machine-readable form for wrappers or agent tooling
- `artifact_workflow_commands.sh --stage <stage> --phase <entry|exit>` returns only the commands for one phase, which is usually the simplest interface for agents

## Core Files
- `flywheel.yaml`
- `flywheel/CONFIG_SCHEMA.md`
- `flywheel/HUMANS.md`
- `flywheel/AGENTS.md`
- `flywheel/DEVELOPMENT_CYCLE.md`

## First-Pass Intent
- Keep the workflow generic.
- Remove product-specific payload.
- Resolve all workflow locations through config.
