# Planning: Cairn Core V1 Rescope

## Objective

Reframe Cairn v1 around a local-first markdown workspace with safe blob-backed sync and local retrieval. The v1 path should be useful to a new engineer without requiring CocoIndex, pgvector, Postgres, a remote indexer service, or remote semantic search.

## Plain-Language Goal

Cairn v1 should let engineers and agents work with real markdown files on disk, validate and index those files locally, and share them through a remote blob store with clear conflict handling. MCP should expose the same safe document operations. Remote indexing and richer retrieval remain future extensions, not required infrastructure for first adoption.

## V1 Product Boundary

In scope for v1:
- Local-first workspace layout with visible markdown files.
- Managed document frontmatter, validation, capture, promote, archive, and CLI-only purge.
- MCP tools for bootstrap, read, find, search, list, validation, lifecycle writes, and safe sync.
- Local SQLite metadata and full-text indexing.
- Remote blob sync through Azure Blob or local filesystem emulation.
- Remote manifest, local sync state, dry-run planning, pull, push, and conflict refusal.
- Developer quickstart and smoke path that prove local-first docs plus blob sync plus local search.

Deferred from v1:
- CocoIndex integration.
- Remote semantic search.
- Remote indexer HTTP service as required infrastructure.
- Postgres/pgvector as required local development services.
- Azure Container Apps indexer deployment and production remote-index auth.
- Generated derived index artifacts as a v1 sync concern.

## Current Findings

- Cairn already has the core document, MCP, local index, remote store, and sync primitives.
- Existing Flywheel artifacts and some current docs still orient the next batch around CocoIndex and a full local service stack.
- The accepted indexing boundary ADR correctly kept Cairn useful without an indexer, but its v1 emphasis now needs to be updated: local metadata/full-text search is the v1 retrieval surface; semantic and remote index modes are deferred adapters.
- The local development harness should prioritize blob sync and local search. A CocoIndex stack can remain reference or future work, but should not define the default v1 success path.

## Proposed Work Items

1. `ARCH-20260507-core-v1-indexing-boundary-refresh`
2. `STORY-20260507-core-v1-docs-and-quickstart`
3. `STORY-20260507-disable-remote-index-mainline`
4. `STORY-20260507-local-blob-sync-smoke`
5. `STORY-20260507-defer-cocoindex-artifacts`

## Recommended Order

1. Refresh the architecture decision boundary so v1 is explicitly local-index-first and CocoIndex is deferred.
2. Update user/product docs and quickstart language to describe the Cairn Core v1 promise.
3. Simplify runtime defaults so local `index_refresh` means local SQLite refresh and `search_context` does not depend on remote services.
4. Ensure the local smoke path proves init, capture, validate, local index/search, sync push, sync pull, and conflict-safe behavior without Postgres/pgvector/CocoIndex.
5. Move CocoIndex and remote indexer artifacts into future/reference positioning so new engineers do not treat them as required v1 setup.

## Success Signals

- A new engineer can explain v1 in five bullets: files, validation, local search, blob sync, MCP.
- The default quickstart does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer.
- Local full-text search works after local index refresh.
- Blob sync works against local filesystem emulation and Azure Blob-compatible storage.
- Conflict handling is documented and covered by the smoke path or tests.
- CocoIndex documentation clearly says deferred or optional rich retrieval.

## Risks

- Existing remote indexer code may continue to imply that v1 requires a service stack unless docs and queue ordering are explicit.
- Removing remote search from the mainline could temporarily reduce perceived ambition; frame it as stabilizing adoption before richer retrieval.
- Some tests or docs may assume `remote_index.url` participates in common flows.
- The local development harness may need a split between core smoke and future rich-retrieval smoke.

## Assumptions

- Local-first visibility is a core adoption feature, not a temporary implementation detail.
- Local SQLite metadata/full-text search is sufficient for the first engineer trial.
- Blob sync plus refusal-based conflict handling is more important for v1 trust than semantic search.
- Keeping the MCP tool contract stable is valuable, even when semantic/remote modes report unavailable.

## Next Stage Recommendation

Next stage: `architect` for `ARCH-20260507-core-v1-indexing-boundary-refresh`.

After the architecture update is accepted, move the engineering stories through PM refinement in the listed order.
