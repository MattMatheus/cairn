# ADR: Indexing Boundary And Query Contract

## Status

Accepted, refreshed for Cairn Core v1 on 2026-05-07.

## Context

Cairn v1 is a local-first markdown workspace for small engineering teams and agents. The core product promise is visible files on disk, managed document lifecycle, validation, local retrieval, MCP access, blob-backed sync, and conflict refusal.

Earlier planning included CocoIndex, pgvector, and a remote indexer in the near-term integration path. Those remain valuable rich-retrieval extensions, but they should not be required for first adoption. A new engineer should be able to run, inspect, and trust Cairn Core v1 without Docker, Postgres, pgvector, CocoIndex, or a remote indexing service.

This ADR defines the v1 indexing boundary and preserves a stable query contract so richer retrieval can return later without forcing MCP or CLI churn.

## Decision

Cairn Core v1 owns and requires:

| Area | V1 owner | Notes |
| --- | --- | --- |
| Document discovery | Cairn core | Walk managed workspace folders, respecting `.cairnignore`. |
| Frontmatter parsing | Cairn core | Durable document ids, lifecycle state, type, tags, authors, actors, and source remain Cairn-managed. |
| Validation state | Cairn core | Managed markdown must pass core schema/lifecycle checks unless ignored. |
| Local metadata index | Cairn core | SQLite index supports document lookup by title, slug, tag, status, type, path, actor, source, and recency. |
| Local full-text search | Cairn core | V1 text retrieval must work with no external indexer. |
| Query contracts | Cairn core | CLI and MCP return stable result/envelope shapes with warnings and unavailable modes. |
| Blob sync boundary | Cairn core | Sync moves workspace documents and control metadata, not generated rich index artifacts. |
| Conflict behavior | Cairn core | Pull/push refuse divergence instead of merging silently. |

CocoIndex and remote semantic indexing are deferred from the v1 critical path:

| Area | V1 status | Notes |
| --- | --- | --- |
| Semantic embeddings | Deferred | Future rich-retrieval adapter behind the existing query contract. |
| Rich summaries/entities/graph features | Deferred | Future CocoIndex-owned derived context, not required for v1. |
| Postgres/pgvector | Deferred | Not required for default local development or v1 smoke. |
| Remote indexer HTTP service | Optional/deferred | May remain as experimental or reference code, but v1 does not depend on it. |
| Production remote-index auth | Deferred | Not needed until a remote indexer is promoted back into the mainline. |

Cairn maintains a lightweight local SQLite index at:

```text
/.cairn/index/cairn.db
```

For v1, `index_refresh` rebuilds or updates this local SQLite index. If a remote indexer is configured in an experimental profile, Cairn may also call it, but a missing or failing remote indexer must not make local indexing fail.

## Search Contract

`search_context` supports these stable modes:

```text
auto
metadata
full_text
semantic
```

V1 mode behavior:

| Mode | V1 behavior |
| --- | --- |
| `metadata` | Query the local SQLite metadata index. |
| `full_text` | Query local full-text content/index state owned by Cairn. |
| `auto` | Attempt local metadata and local full-text. Report any configured-but-unavailable richer modes as degraded/unavailable without treating them as required. |
| `semantic` | Return semantic results only when an optional rich-retrieval adapter is configured and available. Otherwise return a clear unavailable/degraded response with next steps pointing back to local search. |

Search results retain the stable result shape:

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
  "match_type": "full_text",
  "snippet": "Retry auth token requests with bounded exponential backoff...",
  "provenance": {
    "authors": ["matt"],
    "actors": ["codex"],
    "source": "promotion"
  }
}
```

Responses should continue to include attempted modes, unavailable modes, warnings, provenance, and suggested next steps so agents can recover without reading the whole workspace.

## Sync And Index Artifacts

Document sync remains separate from indexing:

- `sync_push` uploads workspace documents and Cairn control metadata to the configured blob store.
- `sync_pull` applies safe remote workspace changes locally and should suggest local `index_refresh` afterward.
- Local `.cairn/index/cairn.db` is generated local state and should not be treated as normal shared document content.
- Generated rich index artifacts are not part of v1 sync.

The remote blob store is the sharing/durability layer for documents, not the v1 query engine.

## Optional Future Remote Indexer Contract

If rich retrieval returns after v1, a remote indexer may expose:

| Endpoint | Responsibility |
| --- | --- |
| `/index/status` | Report optional rich index availability, freshness, and last refresh state. |
| `/index/refresh` | Trigger or request refresh for the configured pod workspace. |
| `/search` | Execute semantic or richer derived-context search behind Cairn's result contract. |

Those endpoints are not required for Cairn Core v1. Production auth, hosted deployment, and pgvector/CocoIndex packaging should be designed when the rich-retrieval adapter is promoted back into active scope.

## Alternatives Considered

- Keep CocoIndex integration in the v1 mainline. Rejected because it makes first adoption depend on too many moving services before engineers have validated the file-first workflow.
- Remove indexing concepts entirely from v1. Rejected because local lookup and full-text retrieval are core to useful MCP/document workflows.
- Make blob storage canonical and drop local-first sync. Rejected because visible local documents are a core trust and adoption feature.
- Split the MCP contract into separate core and rich-retrieval versions immediately. Rejected because the current contract can preserve compatibility through graceful degradation.
- Sync generated index artifacts as normal workspace documents. Rejected because generated artifacts have different freshness and lifecycle semantics from human-authored markdown.

## Consequences

Cairn Core v1 becomes simpler to explain, operate, and validate:

- Engineers need local files, the Cairn binary, and a configured blob store or local filesystem remote-store emulator.
- Search quality is intentionally bounded to local metadata and full text.
- Remote indexer failures are not v1 product blockers.
- The richer CocoIndex path can evolve later behind the existing query contract.

The tradeoff is that v1 does not provide semantic search, entity extraction, graph features, or remote corpus search until those features are explicitly promoted as rich retrieval work.

## Follow-On Implementation Paths

Required for Cairn Core v1:

- Update user/product docs and quickstart around files, validation, local search, blob sync, MCP, and conflict refusal.
- Verify `index_refresh` succeeds locally without `remote_index.url`.
- Verify `search_context` uses local metadata/full-text in `auto` mode without remote services.
- Ensure explicit `semantic` requests degrade clearly when no rich-retrieval adapter is configured.
- Add or update a local blob-sync smoke path that does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer.

Deferred rich-retrieval work:

- Revisit CocoIndex pipeline contracts against the reference repo.
- Reintroduce a remote indexer service as an optional adapter.
- Define pgvector/Postgres packaging if semantic search is promoted.
- Define production remote-index auth and deployment only when the remote indexer is back in scope.
