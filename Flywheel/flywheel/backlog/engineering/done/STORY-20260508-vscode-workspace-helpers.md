# Story: VS Code Workspace Helpers

## Metadata
- `id`: STORY-20260508-vscode-workspace-helpers
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: []
- `success_metric`: A VS Code user can search, capture, promote, validate, and inspect Cairn readiness without leaving the editor.
- `release_scope`: deferred

## Problem Statement

VS Code is common among target developers. A small traditional IDE helper can remove context switching for Cairn workflows without relying on AI IDEs or assistant-specific behavior.

## Scope
- In: Define an initial VS Code extension surface for Cairn commands.
- In: Commands should include capture note, search Cairn, promote current file, validate workspace, and show doctor output.
- In: Reuse the Cairn CLI where practical.
- Out: Cursor or AI IDE integration, rich dashboard UI, autonomous context injection, or source code indexing.

## Assumptions

- The extension can shell out to the Cairn CLI for v1 rather than embedding Cairn logic.
- Repo attachment/workspace discovery should inform how the extension finds the pod KB from a code repo.

## Acceptance Criteria
1. The extension plan identifies commands, workspace discovery behavior, and CLI dependencies.
2. The first implementation can operate from either a Cairn workspace or an attached code repo.
3. Errors from missing CLI, missing workspace, validation failure, or ambiguous workspace are clear and actionable.

## Validation
- Required checks: extension test/lint command once an extension package exists.
- Additional checks: manual VS Code smoke against a sample Cairn workspace.

## Dependencies

- Repo attachment/discovery story.
- Cairn CLI installed or available in PATH.

## Risks

- Extension scope can sprawl into a product UI. Keep v1 command-palette oriented.
- If workspace discovery is weak, editor behavior will feel confusing.

## Open Questions

- Should the extension live in this repository, a sibling package, or a future tools repository?

## Next Step

Engineering should scaffold a small CLI-backed VS Code extension in this repo.

## PM Refinement

- Package location: `extensions/vscode-cairn`.
- First implementation should be JavaScript, command-palette oriented, and dependency-light.
- Commands: capture note, search Cairn, promote current file, validate workspace, and show doctor output.
- Workspace resolution should use an explicit configured workspace path, an opened Cairn workspace, or `.cairn-workspace` discovery through `cairn repo discover`.
- Missing CLI and missing workspace errors must be actionable.

## Engineering Handoff

### What Changed

- Added `extensions/vscode-cairn` as a small JavaScript VS Code extension scaffold.
- Added command-palette commands:
  - `Cairn: Capture Note`
  - `Cairn: Search`
  - `Cairn: Promote Current File`
  - `Cairn: Validate Workspace`
  - `Cairn: Show Doctor`
- Added `cairn.cliPath` and `cairn.workspacePath` settings.
- Workspace discovery supports explicit setting, opened Cairn workspace, and `cairn repo discover --from <folder>`.
- Commands shell out to the Cairn CLI; source code indexing and AI IDE behavior are out of scope.
- Added Node-only tests for workspace discovery, missing CLI messaging, and workspace-relative file promotion paths.

### Validation Evidence

- `npm test` passed in `extensions/vscode-cairn`.
- `go test ./...` passed.
- Node helper smoke verified direct Cairn workspace resolution and repo-pointer discovery behavior.

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, production, or marketplace publishing actions were taken.
- No human approval was required beyond the user's request to continue.

### Risks And Assumptions

- Manual VS Code UI smoke was not run in this environment.
- The first extension is dependency-light JavaScript and intentionally does not package or publish a VSIX.
- Commands depend on a Cairn CLI available via `cairn.cliPath`.
- Promote-current-file only works for files inside the resolved Cairn workspace.

### QA Focus Areas

- Verify extension command manifest includes capture, search, promote current file, validate, and doctor.
- Verify workspace discovery order and missing CLI errors are actionable.
- Verify tests cover both direct Cairn workspace and attached repo discovery.
- Confirm the manual UI smoke gap is recorded as residual risk.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `npm test` passed in `extensions/vscode-cairn`.
- `go test ./...` passed.
- Manifest inspection confirmed commands for capture note, search, promote current file, validate workspace, and doctor.
- Node helper tests cover explicit workspace, opened Cairn workspace, attached repo discovery, missing CLI messaging, and workspace-relative promotion paths.

### Defects

None filed.

### Evidence Quality

Sufficient for a deferred extension scaffold. Manual VS Code host smoke was not run and remains residual risk before packaging or pilot use.

### State Transition

Moved to `done`.
