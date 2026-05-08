# Cairn Pilot Helpers

These helpers reduce pilot friction while keeping Cairn local-first and file-readable.

## Full Readiness Check

Use `doctor --full` when you want one pilot-readiness summary:

```sh
cairn doctor --full
```

It reports config, managed folders, schemas, validation, local index state, search sanity, sync state, remote configuration, and MCP tool surfaces. Remote store reachability is only checked when you ask for it:

```sh
cairn doctor --full --remote
```

Use `--remote` after Azure or `local_fs` sync is configured. Without it, Cairn avoids accidental network or auth calls.

## Low-Friction Capture

For humans who do not want to remember the full capture command:

```sh
cairn note --title "Restart Worker" --type runbook
```

`cairn note` writes a working document under `agents/{actor}/`. The actor defaults from `CAIRN_ACTOR`, `USER`, or `USERNAME`; pass `--actor` when you need an explicit value.

Supported starter types:

```text
note
investigation
handoff
decision
runbook
```

For prompt-driven terminal capture:

```sh
cairn capture --interactive
```

Interactive capture asks for actor, title, type, and markdown body. End the body with a line containing only `.`.

## Multi-Repo Pod Workspaces

Cairn should be one knowledge workspace per pod, even when a pod owns multiple code repos.

Recommended local layout:

```text
/pod-workspace
  /cairn-kb
  /repo-a
  /repo-b
```

Attach repos as references:

```sh
cairn --root ./cairn-kb repo attach --name repo-a --path ../repo-a --url https://dev.azure.com/org/project/_git/repo-a
cairn --root ./cairn-kb repo list
```

By default, `repo attach` writes `.cairn/repos.yaml` in the Cairn workspace and a `.cairn-workspace` pointer in the attached repo. Tools can discover the pod KB from a repo:

```sh
cairn repo discover --from ./repo-a
```

Repo attachment is metadata only. Cairn does not clone, index, sync, or validate repo contents.

## ADO Candidate Capture

The first Azure DevOps lifecycle hook is PR completion. It is designed for pipelines or service hooks that can call the Cairn CLI with a payload file:

```sh
cairn ado capture --event pr-completed --payload-file ado-pr.json
```

This creates a working handoff candidate under `agents/ado/` by default. To send the candidate straight to review staging:

```sh
cairn ado capture --event pr-completed --payload-file ado-pr.json --status proposed
```

`canonical` is intentionally rejected. ADO capture can create candidate knowledge, but humans still promote durable knowledge through Cairn.

The payload parser is fixture-friendly and does not require live ADO auth. It extracts common PR fields such as id, title, description, repository, branches, actor, and URL when present.

Minimal fixture:

```json
{
  "resource": {
    "pullRequestId": 42,
    "title": "Add checkout retry",
    "description": "Retries transient checkout failures.",
    "sourceRefName": "refs/heads/feature/retry",
    "targetRefName": "refs/heads/main",
    "url": "https://dev.azure.com/org/project/_git/payments/pullrequest/42",
    "repository": {
      "name": "payments-api"
    },
    "closedBy": {
      "displayName": "Ada Lovelace"
    }
  }
}
```

## VS Code Helper

The VS Code helper lives under:

```text
extensions/vscode-cairn
```

It is a small command-palette extension that shells out to the Cairn CLI. Commands:

```text
Cairn: Capture Note
Cairn: Search
Cairn: Promote Current File
Cairn: Validate Workspace
Cairn: Show Doctor
```

Settings:

```text
cairn.cliPath
cairn.workspacePath
```

Workspace discovery uses, in order: `cairn.workspacePath`, an open folder containing `.cairn/config.yaml`, or `cairn repo discover` from an attached repo.

Current pilot limit: the extension scaffold has unit tests, but should get a VS Code Extension Host smoke before distribution.

It is not packaged as a VSIX yet. Treat it as a repo-local scaffold until the packaging and Extension Host smoke are done.

## Local Health Report

Generate a local markdown health report:

```sh
cairn health report
```

Write it to a file:

```sh
cairn health report --output .cairn/generated/health.md
```

Prefer `.cairn/generated/` for generated reports. Writing a report under managed folders such as `runbooks/` or `onboarding/` creates ordinary markdown without Cairn document frontmatter, so validation will warn until you deliberately capture or promote a durable summary.

The report includes document counts, proposed documents awaiting review, stale working documents, recent canonical documents, validation findings, local index state, and sync counts. It is descriptive only: no dashboard, scoring, central telemetry, or cross-pod aggregation.
