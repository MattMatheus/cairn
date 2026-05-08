# Cairn VS Code Helpers

This package is a small command-palette extension that shells out to the Cairn CLI. It is intentionally not an AI IDE integration and does not index source code.

## Commands

- `Cairn: Capture Note`
- `Cairn: Search`
- `Cairn: Promote Current File`
- `Cairn: Validate Workspace`
- `Cairn: Show Doctor`

## Workspace Discovery

The extension resolves a Cairn workspace in this order:

1. `cairn.workspacePath` setting.
2. An open VS Code folder that contains `.cairn/config.yaml`.
3. `cairn repo discover --from <folder>` using a `.cairn-workspace` pointer from an attached code repo.

Attached repo contents remain outside Cairn management.
