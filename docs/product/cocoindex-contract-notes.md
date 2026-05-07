# CocoIndex Contract Notes

Status: deferred rich-retrieval reference. CocoIndex is not required for Cairn Core v1. The v1 retrieval path is local SQLite metadata and local full-text search; these notes preserve the future service boundary for semantic search, summaries, entities, and graph features.

## Reference Review

- CocoIndex examples define Python-native flows over sources such as local files, split/chunk content, embed chunks, and write target stores.
- The `code_embedding` example walks local files, chunks markdown/code with `RecursiveSplitter`, embeds with `SentenceTransformerEmbedder`, and stores/query results through Postgres with pgvector.
- The `entire_session_search` example shows metadata plus embedded text rows, reinforcing that Cairn should not depend on one target table shape.
- CocoIndex emphasizes incremental processing and target freshness; Cairn should consume status/search contracts rather than internal artifacts.

## Cairn Contract

Cairn expects an indexer service boundary with:

- `POST /index/status`
- `POST /index/refresh`
- `POST /search`

Responses map into Cairn schema types:

- remote status maps to `mcpschema.IndexStatusData`
- remote search maps to `mcpschema.SearchResult`
- warnings and next steps remain Cairn-owned response metadata

## Follow-Ups

- Local Docker/Podman packaging prototypes were removed from the active tree during the Cairn Core v1 re-scope.
- The active local development direction lives in `deployments/local-dev/README.md`; the historical emulation ADR remains in `docs/adr/ADR-local-development-emulation.md`.
- Defer workspace-to-indexer refresh orchestration until rich retrieval is promoted back into active scope.
- Keep semantic search integration optional until the service endpoint is promoted back into active scope.
