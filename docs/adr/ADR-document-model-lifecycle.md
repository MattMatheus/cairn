# ADR: Document Model And Lifecycle

## Status

Accepted

## Context

Cairn is a portable markdown context layer for small engineering pods. Files remain the source of truth, and markdown must stay readable without Cairn. At the same time, agents need reliable document identity, lifecycle state, and canonical knowledge boundaries so capture, promotion, sync, indexing, and MCP operations behave consistently.

This ADR defines the v1 document model and lifecycle. Privacy filtering, retention policy, hosted SaaS behavior, attachments, and org-wide knowledge features are out of scope for v1.

## Decision

Cairn-managed documents are markdown files that either have a Cairn document id in frontmatter or live under a configured managed folder.

Managed documents should carry core frontmatter:

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

`id` is durable and does not change when a file is renamed or moved. `slug` is human-readable and may change. Filenames are deterministic, lowercase, kebab-case markdown filenames. Time-oriented agent documents may use a date prefix. Inline `#tags` are ignored by Cairn; tags are normalized through frontmatter arrays.

Unknown frontmatter fields produce validation warnings, not failures. Custom schemas may add fields, but must preserve required Cairn core fields.

Cairn is permissive during discovery and strict at durable boundaries:

| Operation | Missing or invalid core frontmatter |
| --- | --- |
| Discover/list ordinary managed markdown | Warning |
| `capture` | Add or repair frontmatter |
| `promote` to `proposed` | Add or repair frontmatter before completing |
| `promote` to `canonical` | Block until valid |
| Sync | Block until valid unless ignored |
| Indexing | Block until valid unless ignored |

Built-in v1 statuses are:

```text
inbox
draft
working
proposed
canonical
archived
```

Allowed transitions are:

```text
inbox -> draft -> proposed -> canonical
working -> proposed -> canonical
anything -> archived
archived -> draft
```

Rejected ideas should normally be archived with:

```yaml
status: archived
archive_reason: rejected
```

Built-in v1 document types are:

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

`note` is the generic fallback type. Workspace config maps document types to destination folders.

Promotion is a first-class Cairn operation exposed through CLI and MCP. Promotion transforms the existing document rather than creating original-versus-promoted copies. It validates or repairs required frontmatter, updates status, moves the file to the configured destination when required, and preserves durable metadata such as `id`, `authors`, `actors`, `source`, `tags`, and timestamps.

Promotion to `proposed` is review staging. Promotion to `canonical` marks durable team knowledge.

Decision documents use ADR-style numbering. ADR numbers are assigned only when a `decision` becomes `canonical`, not when a draft idea is created or merely proposed. Canonical decision promotion moves the document to:

```text
/decisions/ADR-000N-slug.md
```

Archive and purge are separate operations. MCP and interface-level tools may archive documents by setting `status: archived` and moving the document under `/archive`, preserving the original path below the archive root. Hard deletion/purge is CLI-only in v1 and requires explicit confirmation. MCP must not expose hard delete or purge.

## Alternatives Considered

- Require valid frontmatter before any Cairn operation. This improves strictness but makes portable markdown adoption brittle.
- Keep discovery permissive and block only durable boundaries. This is accepted because it preserves markdown portability while protecting canonical, synced, and indexed knowledge.
- Assign ADR numbers when drafts are created. This is rejected because it burns durable numbers for ideas that may never become accepted decisions.
- Copy promoted documents and keep originals. This is rejected because it clutters workspaces and weakens document identity.

## Consequences

Agents can safely discover imperfect markdown, but durable operations have a clear quality gate.

Implementations need a validation layer shared by capture, promotion, sync, indexing, and MCP operations.

ADR numbering needs a deterministic allocation mechanism. If sync can produce concurrent canonical decision promotion, the sync ADR must define how conflicts are detected and refused.

## Follow-On Implementation Paths

- Implement frontmatter parsing and validation.
- Implement document id generation.
- Implement capture and promotion operations.
- Implement archive and CLI-only purge.
- Implement ADR number allocation for canonical decision promotion.
- Add validation tests for lifecycle state transitions and blocking rules.
