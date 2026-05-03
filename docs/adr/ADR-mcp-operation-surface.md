# ADR: MCP Operation Surface

## Status

Accepted

## Context

Cairn gives agents a stable MCP surface for finding, writing, promoting, archiving, syncing, and reading team knowledge. The surface should expose product operations rather than raw filesystem primitives, so agents work through the same validation and lifecycle rules as humans using the CLI.

The v1 surface must stay small enough to support while covering expected agent workflows.

## Decision

Cairn v1 exposes MCP tools for product operations:

| Tool | Mutability | Purpose |
| --- | --- | --- |
| `get_bootstrap` | read | Return compact workspace onboarding context and next steps |
| `capture_note` | local write | Capture agent-authored markdown under `/agents/{actor}/...` |
| `promote_document` | local write | Promote an existing document to a type/status/destination |
| `archive_document` | local write | Archive a document without hard deletion |
| `read_document` | read | Read document metadata, structure, sections, summary, or full text |
| `find_document` | read | Find documents by id, slug, title, path, type, status, or tag |
| `search_context` | read | Search local and optional derived context |
| `list_documents` | read | List managed documents by filters |
| `validate_workspace` | read | Validate managed markdown and sync/index metadata health |
| `sync_status` | read | Report local/remote sync state |
| `sync_pull` | remote write/local write | Pull remote workspace changes when safe |
| `sync_push` | remote write | Push local workspace changes when safe |
| `index_status` | read | Report local/remote index availability and freshness |
| `index_refresh` | local or remote write | Refresh configured index artifacts |

MCP must not expose hard delete or purge in v1.

Tools use a common response shape at decision level:

```json
{
  "ok": true,
  "data": {},
  "warnings": [],
  "unavailable": [],
  "next_steps": [],
  "provenance": {}
}
```

Mutation tools should include changed paths and durable ids in responses. Validation and search tools should include enough metadata for agents to choose a next action without reading entire documents.

Actor identity is required for captures and should be recorded in frontmatter. If an actor is not provided, the MCP server may use a configured default actor for the current agent session.

Profile behavior:

- `local` uses only local workspace state.
- `pod-remote` may use Azure Blob sync and remote indexer calls when configured.

Remote/profile-dependent tools should fail gracefully with warnings and next steps when auth, profile config, or remote services are unavailable.

`search_context` supports:

```text
auto
metadata
full_text
semantic
```

`auto` attempts configured modes in order and reports which modes were attempted, unavailable, or degraded.

`read_document` supports:

```text
summary
frontmatter
toc
sections
full
```

For section reads, agents should request `toc` first, then request specific sections by heading. Full reads are available but should not be the default.

## Alternatives Considered

- Expose raw filesystem primitives. Rejected because it bypasses Cairn validation and lifecycle.
- Collapse all writes into one generic mutation tool. Rejected because explicit operations are safer for agents and easier to validate.
- Keep sync and index operations CLI-only. Rejected because agents need to know when shared context is stale and may need to initiate safe sync/refresh.
- Return full document contents by default. Rejected because it wastes context and weakens progressive disclosure.

## Consequences

The MCP surface becomes a stable contract and should evolve deliberately.

CLI and MCP should share operation implementations where practical.

Agent workflows become safer, but direct filesystem edits remain possible outside Cairn and must be caught by validation.

## Follow-On Implementation Paths

- Define concrete JSON schemas for each MCP tool.
- Implement common response envelope.
- Implement read/search progressive disclosure.
- Implement lifecycle-aware mutation tools.
- Implement profile-aware sync and index tools.
- Add tests for forbidden purge/delete exposure.
