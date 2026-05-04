# ADR: Indexing Boundary And Query Contract

## Status

Accepted

## Context

Cairn should integrate CocoIndex from the start for richer derived context rather than reimplementing semantic indexing, incremental processing, entity extraction, or graph features. Cairn core must still remain useful as a local markdown, MCP, and sync tool when the indexer is disabled or unavailable.

This ADR defines the ownership boundary and v1 query contract. Exact CocoIndex pipeline internals can evolve behind that contract.

## Decision

Cairn core owns:

- document discovery
- frontmatter parsing
- validation state
- local metadata lookup
- local full-text lookup
- CLI and MCP query contracts
- graceful degradation when richer indexes are unavailable

CocoIndex owns richer derived context:

- semantic embeddings
- richer summaries
- entity extraction
- graph features
- incremental processing beyond Cairn's lightweight metadata index
- derived index artifacts

Cairn maintains a lightweight local SQLite metadata index at:

```text
/.cairn/index/cairn.db
```

The local metadata index supports fast lookup by title, slug, tag, status, type, path, actor, source, and recent changes. Local full-text lookup is Cairn-owned for v1 and should not require the external indexer.

Search order for `search_context` is:

1. local metadata
2. local full-text
3. local semantic index when available
4. remote indexer when configured

`search_context` supports modes:

```text
auto
metadata
full_text
semantic
```

`auto` should degrade gracefully and return attempted modes, unavailable modes, warnings, and suggested next steps.

Search results use a stable result shape:

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

The remote indexer should expose endpoint responsibilities:

| Endpoint | Responsibility |
| --- | --- |
| `/index/status` | Report index availability, freshness, and last refresh state |
| `/index/refresh` | Trigger or request refresh for the configured pod workspace |
| `/search` | Execute semantic or richer derived context search behind Cairn's result contract |

The `pod-remote` profile uses Azure CLI identity as the default credential source for remote indexer calls. The remote indexer should validate identity through Azure-managed auth or an equivalent bearer-token flow. Exact ACA auth enforcement remains an ADR follow-up detail within this decision area.

Index artifacts are not normal document sync by default. `sync_push` uploads workspace documents and control metadata. Index refresh publishes or updates derived artifacts separately. `sync_pull` should suggest an index refresh after new changes arrive.

## Alternatives Considered

- Reimplement semantic indexing in Cairn core. Rejected because CocoIndex should power derived context.
- Require CocoIndex for all search. Rejected because Cairn must remain useful without an indexer.
- Sync derived index artifacts as normal workspace documents. Rejected because indexes are generated artifacts with different lifecycle and freshness rules.
- Expose CocoIndex artifact formats directly through MCP. Rejected because Cairn needs a stable query contract independent of internal artifact changes.

## Consequences

Cairn can provide reliable local metadata and full-text behavior even when richer indexes are absent.

CocoIndex can evolve pipelines and artifact formats behind a stable Cairn query contract.

Pods may have uneven search richness depending on local or remote indexer availability, so MCP responses must report degraded modes clearly.

Remote indexer packaging and ACA auth details remain design work under this ADR area, but they do not block the core boundary decision.

## Follow-On Implementation Paths

- Implement local SQLite metadata index.
- Implement local full-text lookup.
- Define concrete query request/response schemas.
- Prototype CocoIndex pipeline contracts against the reference repo.
- Define local Docker/Podman packaging.
- Define Azure Container Apps deployment and auth enforcement.
- Add tests for search degradation and result metadata.
