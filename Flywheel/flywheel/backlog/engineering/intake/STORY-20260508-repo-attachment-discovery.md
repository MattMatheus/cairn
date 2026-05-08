# Story: Repo Attachment And Workspace Discovery

## Metadata
- `id`: STORY-20260508-repo-attachment-discovery
- `owner_role`: Software Architect
- `status`: intake
- `source`: planning
- `decision_refs`: []
- `success_metric`: A multi-repo pod can configure or discover one Cairn workspace from nearby code repositories without duplicating the knowledge base.
- `release_scope`: required

## Problem Statement

Many pods own multiple repositories. Cairn should support one pod knowledge base alongside those repos while avoiding confusion from duplicated workspace state inside every repository.

## Scope
- In: Define a lightweight way to attach repository metadata to a Cairn workspace.
- In: Define workspace discovery from a code repo, likely through a pointer file or nearby workspace search.
- In: Support CLI listing/inspection of attached repos.
- Out: Cloning repositories, indexing source code, validating repo-local docs by default, syncing Cairn docs back into repos, or cross-pod repo discovery.

## Assumptions

- Cairn remains the canonical pod knowledge workspace.
- Repository attachment is metadata and discovery, not content ownership.
- Future VS Code and ADO work can reuse the same repo metadata.

## Acceptance Criteria
1. A Cairn workspace can record attached repo names and paths/URLs in a stable config or managed metadata surface.
2. From a code repo, Cairn tooling can deterministically find the associated pod workspace through an explicit pointer or nearby discovery rule.
3. CLI output makes the one-workspace-many-repos boundary clear.

## Validation
- Required checks: `go test ./...`
- Additional checks: local fixture with one Cairn workspace and two sibling repos.

## Dependencies

- Workspace config model and CLI command surface.

## Risks

- Poorly bounded repo discovery could pick the wrong workspace. Prefer explicit pointers over guessing when ambiguity exists.
- Users may expect source indexing once repos are attached. Documentation and command names should reinforce that attach means reference/discovery only.

## Open Questions

- Should repo metadata live in `.cairn/config.yaml`, a managed `services/` document, or both?
- Should the pointer file be `.cairn-workspace`, `.cairn`, or another name?

## Next Step

PM should decide whether this story needs an architecture pass before implementation.
