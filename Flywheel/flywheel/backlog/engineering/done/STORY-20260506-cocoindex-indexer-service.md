# Story: CocoIndex-Backed Indexer Service

## Metadata
- `id`: STORY-20260506-cocoindex-indexer-service
- `owner_role`: Software Architect
- `status`: done
- `source`: planning
- `decision_refs`: [ADR-indexing-query-boundary, ARCH-20260506-local-azure-emulation-strategy, ADR-local-development-emulation]
- `success_metric`: Remote index refresh and search use CocoIndex + pgvector in the local dev stack.
- `release_scope`: required

## Problem Statement
- The current `cairn-indexer` prototype scans local files and returns deterministic lexical results.
- CocoIndex integration needs an indexer that reads workspace documents from the blob-backed store and stores/query semantic artifacts in Postgres/pgvector.

## Scope
- In:
  - Add a CocoIndex-backed indexer implementation or service.
  - Preserve the existing Cairn HTTP contract.
  - Read managed markdown through the configured remote-store backend.
  - Chunk/embed/index documents into Postgres/pgvector.
  - Return Cairn-shaped search results.
- Out:
  - Production Azure auth enforcement.
  - Non-markdown sources.
  - Multi-pod sharding beyond one configured workspace/container/prefix.

## Assumptions
- Python is acceptable for the CocoIndex service if that is the cleanest integration path.
- `pgvector/pgvector:pg16` is the local database target.
- The initial embedder can use a small local sentence-transformers model.

## Acceptance Criteria
1. `POST /index/status` reports local CocoIndex/Postgres availability and freshness.
2. `POST /index/refresh` indexes managed markdown from blob-backed workspace objects.
3. `POST /search` returns stable Cairn result shape with semantic provenance.
4. Existing Cairn CLI `index status`, `index refresh`, and `search` can exercise the service through `remote_index.url`.
5. The existing Go prototype remains available as a lightweight contract fixture unless the new service fully replaces its smoke-test role.

## Validation
- Required checks:
  - Contract smoke tests for all three HTTP endpoints.
  - CLI smoke test against local `remote_index.url`.
  - Postgres table/vector extension presence check.
- Additional checks:
  - Search after document edit and refresh returns updated content.

## Dependencies
- `ARCH-20260506-local-azure-emulation-strategy`
- `ADR-local-development-emulation`
- `STORY-20260506-dev-blob-store-emulator`
- `STORY-20260506-local-dev-compose-harness`

## Risks
- Model downloads and Python dependencies may make the image large.
- Async refresh semantics may be needed if indexing becomes slow.
- Contract shape may reveal gaps in current schema for semantic score/provenance.

## Open Questions
- Resolved by PM: keep the Go prototype as a lightweight contract fixture until the CocoIndex service fully replaces its smoke-test role.

## Next Step
- QA: verify the local service contract, CLI routing through `remote_index.url`, and regression tests.

## Engineering Handoff
- Added a Python FastAPI CocoIndex service under `deployments/cocoindex-indexer/`, packaged by the local compose stack.
- Local indexer reads managed markdown from Azurite-backed workspace objects, chunks with CocoIndex when available, writes pgvector rows in Postgres, and serves `/index/status`, `/index/refresh`, and `/search`.
- Fixed Azurite SharedKey signing for both the Python service and Go remote-store writes so the CLI can push documents into the emulator.
- Routed CLI `search` through `mcpops.OpenLocal` so semantic mode can use `remote_index.url` instead of staying local-only.

## Engineering Validation
- `python3 -m py_compile deployments/cocoindex-indexer/cairn_cocoindex_indexer/app.py`
- `podman compose -f deployments/local-dev/compose.yml --env-file deployments/local-dev/.env.example up --build -d indexer`
- `curl -sS http://localhost:8080/index/status ...` returned `available: true`.
- `curl -sS http://localhost:8080/index/refresh ...` indexed 4 managed documents from Azurite.
- Temporary Cairn workspace smoke:
  - `cairn sync push` uploaded workspace documents to Azurite.
  - `cairn index status` reported local and remote index availability.
  - `cairn index refresh` refreshed the remote index.
  - `cairn search --mode semantic --query zephyr-lantern` returned the uploaded smoke note.
- `curl -sS http://localhost:8080/search ...` returned `agents/codex/cocoindex-smoke-note.md` with `provenance.source = cocoindex_pgvector`.
- `GOCACHE=$PWD/.cache/go-build go test ./...`

## QA Verdict
- Pass.

## QA Evidence
- Rechecked `POST /index/status`: `available: true`, `fresh: true`, `indexed_count: 4`.
- Rechecked `POST /search` for `zephyr-lantern`: returned `agents/codex/cocoindex-smoke-note.md` and `provenance.source = cocoindex_pgvector`.
- Rechecked CLI:
  - `cairn index status`: local and remote index available.
  - `cairn search --mode semantic --query zephyr-lantern`: returned the smoke note through `remote_index.url`.
- Rechecked regression suite: `GOCACHE=$PWD/.cache/go-build go test ./...`.

## Residual Risk
- The initial embedder is deterministic and local-dev oriented. Replacing it with a production-grade embedding model remains future integration work.
