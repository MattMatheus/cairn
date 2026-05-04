# Story: Parse Generated Pod Remote Profile

## Metadata
- `id`: STORY-20260504-parse-generated-pod-remote-profile
- `owner_role`: Software Architect
- `status`: done
- `source`: qa
- `decision_refs`: [ADR-sync-conflict-behavior, ADR-indexing-query-boundary, ADR-mcp-operation-surface]
- `success_metric`: A workspace initialized by `cairn init` can enable Azure Blob sync and remote indexer access by editing the generated `profiles.pod-remote` config.
- `release_scope`: required

## Problem Statement
- `cairn init` writes Azure settings under `profiles.pod-remote`, but `LoadConfig` only parses top-level `remote_sync` and `remote_index`. Users following the starter config can fill in remote values and still get local-only behavior.

## Scope
- In:
  - Parse generated `profiles.local` and `profiles.pod-remote` config.
  - Map enabled `profiles.pod-remote` values into `RemoteSync` and `RemoteIndex`.
  - Preserve support for existing top-level `remote_sync` and `remote_index` config if needed for compatibility.
  - Validate malformed profile entries with clear findings.
  - Update user docs if the supported config shape changes.
- Out:
  - Secrets or new credential storage.
  - Live Azure resource provisioning.

## Assumptions
- The generated config should be a working template, not merely documentation.

## Acceptance Criteria
1. `LoadConfig` reads remote sync and remote index settings from the generated `profiles.pod-remote` block.
2. `OpenLocal` configures `RemoteStore` and `RemoteIndex` when `pod-remote.enabled` is true and required fields are present.
3. Local-only workspaces remain local-only by default.
4. Validation reports missing required remote fields when `pod-remote.enabled` is true.
5. Tests cover generated config parsing and backward-compatible top-level config parsing.

## Validation
- Required checks:
  - `GOCACHE=/private/tmp/cairn-gocache go test ./internal/document ./internal/workspace ./internal/mcpops`
  - `GOCACHE=/private/tmp/cairn-gocache go test ./...`
- Additional checks:
  - Run `cairn init` in a temp workspace and verify edited generated config activates remote clients.

## Dependencies
- None.

## Risks
- Supporting two config shapes can create precedence ambiguity.

## Open Questions
- Should top-level `remote_sync` and `remote_index` be deprecated immediately or retained silently for early adopters?

## Next Step
- Completed; no follow-up required for this story.

## Engineering Handoff
- `What changed`: Config parsing now supports generated `profiles.pod-remote` remote sync and remote index fields when enabled, while preserving top-level remote config compatibility. Validation reports missing required enabled pod-remote fields.
- `Validation evidence`: `GOCACHE=/private/tmp/cairn-gocache go test ./internal/document ./internal/workspace ./internal/mcpops`; `GOCACHE=/private/tmp/cairn-gocache go test ./...`.
- `Action class`: local write.
- `Approval required`: no.
- `QA focus`: Confirm generated config shape activates remote clients only when `pod-remote.enabled` is true.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: Added parser, validation, and `OpenLocal` tests for generated profile config. Full test suite passed.
- `Evidence quality call`: Sufficient for story acceptance.
- `Defects`: None.
- `State transition decision`: Move to engineering done.
