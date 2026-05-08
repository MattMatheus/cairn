# Story: VS Code Workspace Helpers

## Metadata
- `id`: STORY-20260508-vscode-workspace-helpers
- `owner_role`: Software Architect
- `status`: intake
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

Defer until CLI discovery and capture flows are stable.
