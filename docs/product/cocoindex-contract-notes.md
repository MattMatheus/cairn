# CocoIndex Contract Notes

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

- Define local Docker/Podman packaging for the CocoIndex service.
- Define Azure Container Apps deployment and auth enforcement.
- Define workspace-to-indexer refresh orchestration after sync pull/push.
- Add semantic search integration into `search_context` once the service endpoint exists.

