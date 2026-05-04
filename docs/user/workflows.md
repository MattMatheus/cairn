# Cairn Daily Workflows

## Agent Capture Loop

1. Capture rough work:

   ```sh
   cairn capture --actor codex --title "Thing Learned" --type note --body-file -
   ```

2. Validate the workspace:

   ```sh
   cairn validate
   ```

3. Promote when useful to the pod:

   ```sh
   cairn promote agents/codex/thing-learned.md --type note
   ```

4. Sync when sharing is needed:

   ```sh
   cairn sync dry-run
   cairn sync push
   ```

## Decision Workflow

1. Capture or draft a decision.
2. Promote to `proposed` for review.
3. Promote to `canonical` after acceptance.

```sh
cairn promote working/choose-index-boundary.md --type decision --status canonical
```

Canonical decisions are written as `decisions/ADR-000N-slug.md`.

## Cleanup Workflow

Use archive for normal cleanup:

```sh
cairn archive runbooks/old-runbook.md
```

Use purge only after archive when the local file must be hard-deleted:

```sh
cairn purge --confirm-purge archive/runbooks/old-runbook.md
cairn sync dry-run
```

The dry-run should show a delete if the archived file had already been synced.

## Ignore Local Noise

Put local-only files in `.cairnignore`:

```gitignore
.DS_Store
.cairn/index/
.cairn/generated/
scratch/
*.tmp
```

Ignored paths are skipped by validation, indexing/search, and sync manifests.

## Healthy Workspace Checklist

- `cairn validate` has no errors.
- `cairn index refresh` completes locally.
- `cairn search "known term"` finds expected documents.
- `cairn sync dry-run` is clean or shows expected pull/push changes.
- `AGENTS.md` and `CLAUDE.md` stay short and point agents to onboarding docs.

## Remote Profile Checklist

Before relying on remote sync or indexing:

- Azure CLI login works for the pod tenant.
- `.cairn/config.yaml` has `remote_sync` and `remote_index` values.
- No secrets are stored in workspace config.
- `cairn sync status` reports remote state without auth failures.
- `cairn index status` reports remote availability or a clear degraded mode.
