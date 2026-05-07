# Pilot Guide

This guide is the first-run path for a friendly engineering pilot. The goal is to prove that Cairn is useful as a local-first markdown workspace before asking anyone to care about richer retrieval or hosted infrastructure.

## Pilot Goal

A pilot engineer should be able to:

- install a released Cairn binary into userland
- initialize a fresh workspace with `cairn init`
- confirm `.cairn/config.yaml` exists
- configure no-service sync with `cairn setup local-sync`
- inspect visible markdown documents
- capture a new document
- validate the workspace
- refresh the local SQLite index
- find documents with local search
- push and pull through a no-service `local_fs` remote store
- see that conflicts are refused rather than silently merged

No Docker, cloud account, Postgres, pgvector, CocoIndex, Azurite, or remote indexer is required.

## Maintainer Preflight

Run this before inviting a pilot:

```sh
make pilot-check
```

Expected final line:

```text
Pilot check passed.
```

This runs the Go test suite, builds a throwaway binary, validates and searches the example workspace, and runs the no-service sync smoke.

## Pilot Setup

Install Cairn into userland. Pilots should not need the Go toolchain or a repository-local `bin/` shim:

```sh
curl -fsSL https://raw.githubusercontent.com/MattMatheus/cairn/main/scripts/install.sh | sh
export PATH="$HOME/.local/bin:$PATH"
cairn version
```

Expected: Cairn prints its version. For unreleased internal builds, install the attached pilot binary to `~/.local/bin/cairn` and run the same `cairn version` check.

Do not run `cairn init` inside the Cairn source repository. Use `--root` with a normal workspace directory as shown below.

## Initialize A Fresh Workspace

Create and initialize a throwaway workspace:

```sh
WORK_ROOT="$(mktemp -d)/cairn-pilot"
cairn --root "$WORK_ROOT" init --workspace-id cairn:workspace:pilot
test -f "$WORK_ROOT/.cairn/config.yaml"
cairn --root "$WORK_ROOT" doctor
```

Expected:

```text
Initialized workspace cairn:workspace:pilot
```

The `test -f` command should print nothing and exit successfully. If `.cairn/config.yaml` is missing, stop there; sync will not be configured correctly.

## Configure No-Service Sync

Run the setup helper so you do not have to edit YAML by hand:

```sh
cairn --root "$WORK_ROOT" setup local-sync --remote-root /tmp/cairn-pilot-remote
cairn --root "$WORK_ROOT" doctor
```

Expected:

```text
Configured local sync in .cairn/config.yaml
```

This command is safe to run on a new workspace. It creates the standard Cairn files if they do not exist and writes the equivalent of:

```yaml
remote_sync:
  provider: local_fs
  root: /tmp/cairn-pilot-remote
```

## Inspect The Example Workspace

Copy the sample workspace to a throwaway location:

```sh
SAMPLE_ROOT="$(mktemp -d)/cairn-pilot-example"
cp -R "$CAIRN_REPO/examples/pilot-workspace" "$SAMPLE_ROOT"
cd "$SAMPLE_ROOT"
```

Open these files in an editor:

```text
runbooks/pilot-handshake.md
working/sync-expectations.md
onboarding/team-context.md
```

The frontmatter is part of the product surface. It is meant to be readable and reviewable, not hidden in a service.

## Validate And Search

```sh
cairn validate
cairn index status
cairn index refresh
cairn search pilot-handshake
```

Expected:

```text
Workspace validation passed.
Local index refreshed: true
Found 1 result(s).
```

The result should include `runbooks/pilot-handshake.md`.

## Capture A New Note

```sh
cairn capture \
  --actor "$USER" \
  --title "Pilot First Impression" \
  --type note \
  --tags pilot,feedback \
  --body "Cairn should make local context easier to inspect and share."
```

Expected: Cairn prints the created path and next steps. The new document should be a normal markdown file under `agents/<actor>/`.

Run:

```sh
cairn validate
cairn search "first impression"
```

## Try No-Service Sync

The sample workspace already has a config file. Configure a local filesystem remote store with the same setup helper:

```sh
cairn setup local-sync --remote-root /tmp/cairn-pilot-example-remote
```

Then run:

```sh
cairn sync status
cairn sync dry-run
cairn sync push
```

Expected:

```text
Sync diverged: false
Sync dry-run direction: push
Sync push applied: true
```

## Conflict Expectation

For the first pilot, do not ask the engineer to resolve a real conflict manually. Instead, show the automated smoke:

```sh
"$CAIRN_REPO/deployments/local-dev/core-smoke.sh"
```

Expected final line:

```text
Core smoke passed.
```

That script creates a deterministic local/remote divergence and verifies that Cairn reports `Sync diverged: true` plus a `Conflict:` line.

## What To Report

Ask the pilot engineer to write down:

- the first command that felt unclear
- any output that made them hesitate
- whether the markdown files felt trustworthy
- whether search found what they expected
- whether sync behavior was understandable
- what they expected Cairn to do that it did not do

## Known Pilot Limits

- Search is local metadata and full text, not semantic retrieval.
- `local_fs` is a no-service blob-style remote store for pilot testing, not production storage.
- Azure Blob is the multiplayer backend path. It is documented in [Azure Blob Sync](azure-sync.md), but is not required for the first no-service pilot script.
- Remote indexer and CocoIndex work are deferred until the core workflow survives pilot feedback.
