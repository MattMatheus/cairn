# Local Indexer Prototype

This package runs a local HTTP indexer that matches Cairn's remote index contract:

- `POST /index/status`
- `POST /index/refresh`
- `POST /search`

It is a CocoIndex packaging placeholder, not the production pipeline. The service scans managed Cairn markdown and returns deterministic lexical results so smoke tests do not need embeddings, paid model calls, Postgres, or pgvector.

## Build

```sh
docker build -f deployments/local-indexer/Dockerfile -t cairn-indexer:local .
```

Podman works with the same command:

```sh
podman build -f deployments/local-indexer/Dockerfile -t cairn-indexer:local .
```

## Run

```sh
docker run --rm -p 8080:8080 \
  -e CAIRN_WORKSPACE_ROOT=/workspace \
  -v "$PWD:/workspace:ro" \
  cairn-indexer:local
```

Environment variables:

- `CAIRN_WORKSPACE_ROOT`: mounted Cairn workspace path inside the container. Defaults to `/workspace`.
- `CAIRN_INDEXER_ADDR`: listen address. Defaults to `:8080`.

Mounts:

- `/workspace`: read-only Cairn workspace containing managed markdown.

## Smoke

```sh
curl -s http://localhost:8080/index/status \
  -H 'content-type: application/json' \
  -d '{"workspace_id":"local"}'

curl -s http://localhost:8080/index/refresh \
  -H 'content-type: application/json' \
  -d '{"workspace_id":"local"}'

curl -s http://localhost:8080/search \
  -H 'content-type: application/json' \
  -d '{"workspace_id":"local","query":"auth","limit":5}'
```

Expected behavior:

- Status returns `available: true`.
- Refresh returns `accepted: true` and `refreshed: true`.
- Search returns Cairn-shaped results for matching managed markdown files.
