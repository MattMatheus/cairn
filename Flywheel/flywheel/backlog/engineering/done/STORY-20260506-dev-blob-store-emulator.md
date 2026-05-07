# Story: Dev Blob Store Emulator

## Metadata
- `id`: STORY-20260506-dev-blob-store-emulator
- `owner_role`: Database Expert
- `status`: done
- `source`: planning
- `decision_refs`: [ARCH-20260506-local-azure-emulation-strategy, ADR-sync-conflict-behavior, ADR-local-development-emulation]
- `success_metric`: Cairn sync status, dry-run, push, and pull work against local filesystem storage and Azurite.
- `release_scope`: required

## Problem Statement
- Remote sync currently targets Azure Blob REST paths and bearer-token auth.
- Local developers need both a frictionless local filesystem remote-store mode and an Azure Blob-compatible Azurite mode without Azure CLI login, storage accounts, or cloud resources.

## Scope
- In:
  - Implement a `local_fs` remote-store backend.
  - Configure Azurite as the Azure Blob-compatible emulator path.
  - Add any Cairn config or token-provider changes needed for local dev and Azurite auth.
  - Support remote manifest and object read/write/delete/list operations used by Cairn.
  - Document local config values for `.cairn/config.yaml`.
- Out:
  - Full Azure Blob API coverage.
  - Production auth.
  - Multi-tenant access control.

## Assumptions
- A static dev token, no-op token path, or Azurite development credentials are acceptable when the endpoint is local.
- Azurite is the full integration path; `local_fs` is the quick inner-loop path.

## Acceptance Criteria
1. `cairn sync status` can read local remote manifest state from `local_fs`.
2. `cairn sync push` writes workspace objects and `.cairn/remote-manifest.json` to `local_fs`.
3. `cairn sync pull` can restore remote-only changes from `local_fs` into a workspace.
4. The same sync smoke path works against Azurite through `azure_blob` with explicit `auth_mode: azurite` configuration.
5. Missing or divergent manifests still produce the same safety behavior as the existing sync model.

## Validation
- Required checks:
  - Local sync push/pull smoke test against `local_fs`.
  - Local sync push/pull smoke test against Azurite.
  - Existing sync unit tests pass.
- Additional checks:
  - Exercise paths with spaces and prefixes.

## Dependencies
- `ADR-local-development-emulation`
- `STORY-20260506-local-dev-compose-harness`

## Risks
- Two store modes increase the surface area.
- Azurite may require auth behavior not currently modeled by Cairn.

## Open Questions
- Resolved by PM: use explicit `auth_mode: azurite` in Cairn config; environment variables may provide default Azurite credentials for scripts/containers.

## Next Step
- QA: verify `local_fs` sync behavior, Azurite auth/config support, and regression coverage.

## Engineering Handoff
- Action model: local write.
- Approval: not required; no risky, sensitive, or production action taken.
- Changed implementation/docs:
  - Added `internal/remotestore/local_fs.go`.
  - Extended `RemoteSyncConfig` with `root` and `auth_mode`.
  - Wired `local_fs` and Azurite auth mode in `mcpops.OpenLocal`.
  - Added Azurite SharedKey authorization path for `azure_blob` with `auth_mode: azurite`.
  - Updated local-dev compose/env/docs with `AZURITE_ACCOUNT_KEY` and Cairn config snippets.
  - Added tests for config parsing, client wiring, `local_fs`, and Azurite auth headers.
- Validation:
  - `GOCACHE=$PWD/.cache/go-build go test ./internal/remotestore ./internal/document ./internal/mcpops`
  - `GOCACHE=$PWD/.cache/go-build go test ./...`
  - CLI smoke against `local_fs`: initialized workspace A, configured external `/private/tmp/...` local remote root, pushed, copied synced base to workspace B, captured/pushed a second note from A, verified B saw one remote create, then pulled it successfully.
- Validation result:
  - Targeted and full Go test suites passed.
  - `local_fs` push and pull smoke passed.
  - Azurite runtime smoke could not run because Docker is not installed and Podman socket access is blocked in this environment.
  - Azurite request auth behavior is covered by `TestAzureBlobStoreAzuriteAuthUsesSharedKey`.
- Open risks:
  - Full Azurite end-to-end sync should be run on a developer machine with Docker/Podman.
  - The initial Azurite key is local-dev only and intentionally configured through `.env.example`.
- QA focus:
  - Confirm absolute `local_fs` roots stay outside the workspace manifest.
  - Confirm `auth_mode: azurite` does not invoke Azure CLI token acquisition.
  - Confirm existing Azure Blob bearer-token behavior still works.

## QA Verdict
- Verdict: pass with Azurite runtime limitation recorded.
- Evidence summary:
  - Targeted tests passed for `internal/remotestore`, `internal/document`, `internal/mcpops`, and `internal/syncstate`.
  - Full `go test ./...` passed.
  - Source/test scan confirms `local_fs`, `auth_mode`, Azurite auth, Azure CLI bearer-token behavior, and config snippets are present.
  - Engineering local_fs CLI smoke demonstrated push and pull through an external local remote root.
- Defects: none filed.
- Required fixes: none.
- Residual risk: full Azurite sync smoke still needs Docker/Podman access on a developer machine.
- Next state recommendation: done.
