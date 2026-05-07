# Story: Core V1 Docs And Quickstart

## Metadata
- `id`: STORY-20260507-core-v1-docs-and-quickstart
- `owner_role`: Technical Writer
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: A new engineer can complete the default quickstart without CocoIndex, Postgres, pgvector, Docker, or a remote indexer.
- `release_scope`: required

## Problem Statement
- Current docs still emphasize the CocoIndex/local service stack path. V1 docs should present Cairn Core as local markdown files, validation, local search, MCP, and safe blob sync.

## Scope
- In:
  - Update README, user quickstart, workflows, north-star/product notes, and local development docs as needed.
  - Make the default setup path local-first and local-index-first.
  - Mark CocoIndex and remote semantic search as deferred or optional rich retrieval.
  - Make the local filesystem remote-store path the primary no-service sync example, with Azure Blob/Azurite as secondary integration paths.
- Out:
  - Code changes to search, sync, or index behavior.
  - Removing reference material unless it is misleading in the default path.

## Assumptions
- Architecture refresh has clarified the v1 boundary before this story starts.
- Existing local index and sync commands remain part of the user-facing surface.

## Acceptance Criteria
1. Default docs describe the v1 promise as files, validation, local search, blob sync, and MCP.
2. Quickstart does not require CocoIndex, Postgres, pgvector, Docker, or a remote indexer.
3. CocoIndex/remote indexer docs are labeled deferred, optional, or future rich retrieval.
4. Sync conflict behavior remains visible in user docs.
5. Docs include a short path for local filesystem remote-store emulation.

## Validation
- Required checks:
  - Documentation review for stale statements that make CocoIndex required for v1.
  - Run or cite existing doc-related smoke commands when available.
- Additional checks:
  - Ask a reviewer to follow the default quickstart from a clean workspace.

## Dependencies
- `ARCH-20260507-core-v1-indexing-boundary-refresh` is accepted.

## Risks
- Some existing docs may intentionally describe future infrastructure; preserve them as future/reference material instead of deleting useful context.

## Open Questions
- None for activation. PM selected local filesystem remote-store as the primary docs path, with Azure Blob/Azurite as secondary.

## PM Handoff
- `What changed`: Promoted into active as the first Cairn Core v1 story.
- `Why it matters`: The v1 target needs to be visible before runtime and smoke changes are validated.
- `Acceptance criteria`: Kept from intake and tightened to make local filesystem remote-store primary.
- `Risks and assumptions`: Preserve useful future rich-retrieval docs while removing default-path confusion.
- `Next state recommendation`: Engineering should implement this first and move to QA with documentation review evidence.

## Engineering Handoff
- `Docs updated`: Updated `README.md`, `docs/user/README.md`, `docs/user/quickstart.md`, `docs/user/workflows.md`, `deployments/local-dev/README.md`, `docs/product/north-star.md`, and `docs/product/cocoindex-contract-notes.md`.
- `Key clarifications`: Cairn Core v1 is local-first; default search/indexing is local SQLite metadata and full text; default sync documentation uses `remote_sync.provider: local_fs`; Azure Blob/Azurite are secondary integration paths; CocoIndex, pgvector, Postgres, remote semantic search, and the remote indexer are optional/deferred rich retrieval.
- `Removed stale content`: Replaced default-path language that pointed new users at the CocoIndex/local service stack with no-service local-first guidance. Preserved rich-retrieval material as optional/reference instead of deleting it.
- `Validation`: Ran documentation stale-language review with `rg` for required/default CocoIndex, Postgres, pgvector, Docker, and remote-indexer phrasing. Ran `go run ./cmd/cairn help` successfully after sandbox escalation allowed Go build-cache access.
- `Action and approval notes`: Local documentation edits are `local write`. The `go run ./cmd/cairn help` validation initially failed because the sandbox blocked Go build-cache writes under the user cache directory; reran with approved escalation for `go run`.
- `Remaining gaps`: The next active stories still need to verify runtime local-only behavior and add a no-service local blob-sync smoke path.
- `QA focus areas`: Confirm the quickstart default path does not require Docker, Postgres, pgvector, CocoIndex, or a remote indexer; confirm sync conflict behavior remains visible; confirm rich-retrieval references are clearly optional/deferred rather than required.
- `Next state recommendation`: Move to engineering QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Reviewed `README.md`, `docs/user/quickstart.md`, `docs/user/workflows.md`, `deployments/local-dev/README.md`, `docs/product/north-star.md`, and `docs/product/cocoindex-contract-notes.md` against the five acceptance criteria. The docs state the Cairn Core v1 promise, make the quickstart no-service and local-first, label rich retrieval as optional/deferred, keep sync conflict refusal visible, and include a `local_fs` remote-store example.
- `Evidence quality call`: Sufficient for a documentation story. QA also reran targeted `rg` checks for default/required CocoIndex, Postgres, pgvector, Docker, and remote-indexer language, plus `go run ./cmd/cairn help`.
- `Defects`: None.
- `Required fixes`: None.
- `Residual risks`: Runtime behavior and no-service smoke coverage remain in the next active stories.
- `Next state recommendation`: Move to engineering done.

## Next Step
- Continue engineering with `STORY-20260507-disable-remote-index-mainline`.
