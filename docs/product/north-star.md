# Cairn North Star

## Product Promise

Cairn is shared markdown context for humans and agents without app lock-in.

It is a per-pod knowledge and context layer for small engineering teams. Humans can keep using normal markdown tools such as Obsidian, Tolaria, VS Code, Zed, or a plain filesystem. Agents get a stable MCP surface for finding, writing, promoting, and syncing team knowledge.

Cairn is not a markdown editor, a cloud-drive sync client, or a cross-org knowledge graph. It is the coordination layer that makes a portable markdown document store usable by agentic engineering pods.

## First Users

Cairn is built first for small engineering pods of roughly 2-4 people operating in an enterprise environment. The design assumes teams vary in agent maturity, so the system should provide enough structure to make agent context reliable without forcing every team into one development methodology.

The initial internal use case is supporting agentic pod engineering work, where agents need to get up to speed quickly from shared team knowledge and then leave useful notes behind.

## Core Principles

- Files are the source of truth.
- Markdown remains portable and readable without Cairn.
- Cairn is lightly opinionated where consistency helps agents.
- Teams can customize outside the core schema and folder model.
- Agents may write directly, but promotion into canonical knowledge is explicit.
- Permanent deletion is human-controlled through CLI-only purge commands.
- Per-pod isolation is a product boundary for v1.
- Cloud drive sync is out of scope.
- CocoIndex should power derived context rather than be reimplemented.

## Product Layers

Cairn has three conceptual layers:

1. Document store
   Markdown files, frontmatter, folder layout, ignore rules, and sync metadata.

2. Cairn control plane
   Workspace config, schemas, document lifecycle, MCP tools, actor identity, promotion, validation, bootstrap, and Azure Blob sync.

3. Derived context engine
   CocoIndex-backed indexing for search, semantic retrieval, summaries, graph extraction, and future richer context features.

The control plane must not depend on every derived index being available. Cairn should remain useful as a local markdown/MCP/sync tool even when the indexer is disabled.

## Workspace Scope

A Cairn workspace is scoped to one pod.

Cross-pod discovery, central org-wide knowledge search, and knowledge federation are explicitly out of scope for Cairn v1. If those become important, they should be handled by a separate tool that consumes promoted pod knowledge.

## Default Workspace Layout

`cairn init` creates the standard layout:

```text
/inbox
/agents
/working
/decisions
/runbooks
/projects
/services
/handoffs
/onboarding
/archive
/.cairn
```

Default meanings:

- `/inbox`: untrusted or unreviewed inbound content, future imports, and future web clipper output.
- `/agents`: free-use space for agent-authored material, generated bootstrap files, investigations, summaries, and handoffs.
- `/working`: human/team drafts and in-progress notes.
- `/decisions`: canonical decision records using ADR-style numbering.
- `/runbooks`: operational procedures.
- `/projects`: project briefs, plans, and status documents.
- `/services`: service-level documentation.
- `/handoffs`: transition and session handoff documents.
- `/onboarding`: compact team and agent setup documents.
- `/archive`: archived managed documents, preserving original path below `/archive`.
- `/.cairn`: config, schemas, sync state, generated machine artifacts, and indexes.

All top-level Cairn folders are managed by default except ignored paths and `.cairn` internals. A markdown file is Cairn-managed if it either has a Cairn document id or lives under a configured managed folder.

## Starter Onboarding Files

`cairn init` creates these files if they do not exist:

```text
AGENTS.md
CLAUDE.md
/onboarding/team-context.md
/onboarding/agent-setup.md
/onboarding/workspace-map.md
```

`AGENTS.md` and `CLAUDE.md` should be minimal pointers to Cairn tools, not full team constitutions.

## Document Frontmatter

Every Cairn-managed document should have core frontmatter:

```yaml
id: cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A
schema_version: 1
title: Example Document
slug: example-document
type: note
status: working
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors:
  - matt
actors:
  - codex
source: capture
tags: []
```

`id` is durable and should not change when a file is renamed. `slug` is human-readable and may change. Filenames are deterministic, lowercase, kebab-case markdown filenames. Time-oriented agent documents may use a date prefix.

Inline `#tags` are ignored by Cairn. Tags are normalized through frontmatter arrays.

Unknown frontmatter fields should produce validation warnings, not failures.

## Status Model

Built-in statuses:

```text
inbox
draft
working
proposed
canonical
archived
```

Allowed transitions:

```text
inbox -> draft -> proposed -> canonical
working -> proposed -> canonical
anything -> archived
archived -> draft
```

Rejected ideas should normally be archived with an archive reason instead of introducing a separate `rejected` status:

```yaml
status: archived
archive_reason: rejected
```

## Document Types

Built-in v1 document types:

```text
note
handoff
investigation
decision
runbook
project
service
onboarding
```

`note` is the generic fallback type.

Built-in schemas may be stricter. Custom schemas should be YAML files under:

```text
/.cairn/schemas/*.yaml
```

Custom schemas may be permissive, but must include required Cairn core fields.

Workspace config maps document types to destination folders.

## ADRs

Decision documents use ADR-style numbering.

ADR numbers are assigned when a decision is promoted/completed, not when a draft idea is first created. Promotion moves the document to:

```text
/decisions/ADR-000N-slug.md
```

This avoids burning ADR numbers for ideas that never become durable decisions.

## Agent Capture

Agents may write directly through Cairn.

Default agent captures land under:

```text
/agents/{actor}/...
```

The default captured document type is `note` unless the type is obvious or explicitly supplied. Hooks and MCP tools should use first-class Cairn operations rather than writing arbitrary files directly.

Example commands:

```text
cairn capture --actor codex --type investigation --title "Debug auth timeout"
cairn capture --actor claude --type handoff --stdin
```

Capture output should include the created path and suggested next actions, such as promoting the document or syncing the workspace.

## Promotion

Promotion is a core Cairn operation exposed through CLI and MCP.

Promotion should:

- ask for or receive a document type
- validate or add required frontmatter
- update status
- assign ADR number when applicable
- move the document to the configured destination folder when required
- preserve durable metadata such as id, authors, actors, source, tags, and timestamps

Promotion transforms the existing document. It should not clutter the workspace with separate original-versus-promoted copies.

## Archive And Purge

Archive and purge are separate operations.

MCP and interface-level tools may archive documents. Archiving sets:

```yaml
status: archived
```

and moves the document under `/archive`, preserving the original path below the archive root:

```text
/archive/decisions/ADR-0001-old-choice.md
/archive/agents/codex/2026-05-02-investigation.md
```

Hard deletion/purge is CLI-only and should require explicit confirmation. Agents should not receive a hard-delete MCP tool in v1.

## Sync

Cairn v1 uses on-demand sync initiated by a human or agent through CLI or MCP.

Azure Blob storage is the first shared backend. Each pod is expected to have its own Azure Storage Account for isolation, billing clarity, and simpler operations.

Cloud drive sync, including OneDrive, is out of scope.

The remote Blob layout should mirror the local workspace root, optionally under a configured prefix:

```yaml
remote:
  provider: azure_blob
  container: cairn
  prefix: ""
```

Sync operates on the whole workspace except ignored paths, similar to Git. Ignore rules live in:

```text
.cairnignore
```

and should use gitignore-style syntax.

Sync metadata:

- local sync state lives in `/.cairn/sync-state.json`
- remote manifest lives in Blob at `/.cairn/remote-manifest.json`
- the manifest records path, size, hash, modified time, and document id when available

If local and remote both changed since the last known base, Cairn warns and refuses rather than trying to merge.

Archive and move operations are exposed through MCP. Purge/delete is CLI-only.

## Auth

Cairn should assume Azure CLI login for v1.

The CLI may shell out to:

```text
az account get-access-token
```

to obtain bearer tokens. This aligns with existing Azure DevOps skill usage and avoids requiring VPN-only access patterns.

No secrets should be stored in `.cairn/config.yaml`.

## Profiles

The workspace config supports two initial profiles:

```text
local
pod-remote
```

`local` is for local-only operation. `pod-remote` enables Azure Blob sync and remote indexer access.

MCP tools may use the `pod-remote` profile when configured.

## Indexing

Indexing is optional but strongly recommended.

Cairn should integrate CocoIndex from the start through an indexer boundary rather than reimplementing semantic indexing, incremental processing, entity extraction, or graph features.

The core Cairn binary should remain useful without an indexer. When available, the indexer provides richer derived context.

Initial indexing architecture:

- Cairn core is a self-contained Go CLI/MCP binary.
- A local indexer may run in Docker or Podman.
- An Azure Container Apps indexer should be part of the early architecture.
- v1 may run one indexer per pod to simplify isolation, locking, cost attribution, and operations.
- Index outputs live in the same pod storage account.
- Indexes remain pod-scoped; org-wide indexing is out of scope.

Cairn maintains a lightweight local SQLite metadata index at:

```text
/.cairn/index/cairn.db
```

for fast title, slug, tag, status, type, path, actor, source, and recent-change lookups.

CocoIndex owns richer derived index artifacts through its pipelines. Cairn should define stable query contracts rather than depending on every artifact format directly.

The remote indexer should expose HTTP endpoints such as:

```text
/index/status
/index/refresh
/search
```

Search order:

1. local metadata
2. local full-text
3. local semantic index when available
4. remote indexer when configured

`sync_pull` should suggest an index refresh after new changes arrive.

Index artifacts are not normal document sync by default. `sync_push` uploads workspace documents and control metadata. Index refresh publishes or updates index artifacts separately.

## Search And Read Tools

MCP tools should expose product operations rather than raw filesystem primitives.

Initial tool candidates:

```text
get_bootstrap
capture_note
promote_document
archive_document
read_document
find_document
search_context
list_documents
validate_workspace
sync_status
sync_pull
sync_push
index_status
index_refresh
```

`search_context` supports modes:

```text
auto
metadata
full_text
semantic
```

`auto` should degrade gracefully and return which modes were attempted, which were unavailable, warnings, and suggested next steps.

Search results should include enough metadata for agents to identify and choose documents:

```json
{
  "path": "/runbooks/auth-timeouts.md",
  "title": "Auth Timeout Runbook",
  "type": "runbook",
  "status": "canonical",
  "slug": "auth-timeout-runbook",
  "tags": ["auth", "timeouts"],
  "updated": "2026-05-02T12:00:00Z",
  "score": 0.91,
  "match_type": "semantic",
  "snippet": "Retry auth token requests with bounded exponential backoff...",
  "provenance": {
    "authors": ["matt"],
    "actors": ["codex"],
    "source": "promotion"
  }
}
```

`read_document` supports modes:

```text
summary
frontmatter
toc
sections
full
```

For section reads, an agent should request the table of contents first, then request specific sections by heading.

## Bootstrap

Bootstrap is both a generated visible markdown file and an MCP response.

Visible generated files may live under `/agents`. Machine/cache artifacts may live under `/.cairn/generated`.

The bootstrap response should be compact and progressively disclose next steps. It should not burn large context budgets.

Bootstrap should include:

- pod purpose
- workspace layout
- key onboarding docs
- how to query Cairn next
- references to external tools or skills, especially Azure DevOps
- missing setup warnings

Required external skills are configured with name, setup doc, and source link:

```yaml
required_skills:
  - name: azure-devops
    setup_doc: /onboarding/azure-devops.md
    source: https://example.invalid/central-skills-repo
```

If required items are missing, Cairn should prompt the user or agent to reference the workspace setup document.

## Validation

Cairn provides:

```text
cairn validate
validate_workspace
```

Validation checks Cairn-managed markdown only.

Validation severities:

```text
error
warning
info
```

Capture and promotion may auto-add missing frontmatter. Validation should warn on missing or invalid frontmatter because that usually means content bypassed Cairn.

Validation checks should include:

- required frontmatter
- schema version
- known status
- known or configured document type
- deterministic filename
- normalized tags
- destination folder conventions
- ignored unsupported attachments
- sync metadata health where applicable

## Init

`cairn init` should support interactive setup and non-interactive flags.

Interactive setup should create:

- standard folder layout
- `.cairn/config.yaml`
- `.cairnignore`
- starter schemas
- starter onboarding files
- minimal `AGENTS.md`
- minimal `CLAUDE.md`

Non-interactive setup should be available for automation.

## Enterprise Constraints

V1 constraints:

- Windows support is required.
- A self-contained Go binary is preferred for Cairn core.
- Claude Code, Codex, and GitHub Copilot agent flows should be supported.
- Azure CLI login is acceptable and expected.
- No secrets feature in v1.
- Retention policy is v2 or later.
- Privacy filtering for PII leakage prevention is v2 or later.

Future privacy work may integrate a pre-index or pre-promotion filter similar in spirit to OpenAI Privacy Filter so sensitive data can be kept out of the store when possible.

## Explicit Non-Goals For V1

- Markdown editor UI.
- Obsidian-specific required workflow.
- OneDrive or cloud-drive sync.
- Cross-pod search.
- Org-wide knowledge graph.
- Conflict resolution or merge tooling.
- Attachments.
- Secret storage.
- Retention policy.
- MCP hard delete.
- Mobile app.
- Hosted SaaS business model.

## Open Questions

- Exact CocoIndex pipeline contracts and artifact formats.
- Local indexer packaging details.
- ACA indexer deployment and auth details.
- Azure Blob manifest schema.
- YAML schema format for built-in and custom document types.
- Initial command and MCP API schemas.
- How much Git compatibility or nudging should Cairn provide.
- When privacy filtering enters the lifecycle: capture, promotion, sync, or index.
