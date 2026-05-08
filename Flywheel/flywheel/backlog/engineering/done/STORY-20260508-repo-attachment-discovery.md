# Story: Repo Attachment And Workspace Discovery

## Metadata
- `id`: STORY-20260508-repo-attachment-discovery
- `owner_role`: Software Architect
- `status`: done
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

Engineering should implement explicit repo metadata plus pointer-file discovery without source indexing or repo ownership.

## PM Refinement

- Use `.cairn/repos.yaml` for attached repo metadata.
- Use `.cairn-workspace` as the explicit repo-local pointer file.
- Provide `cairn repo attach`, `cairn repo list`, and `cairn repo discover`.
- Keep output clear that attached repos are references only; Cairn does not clone, index, sync, or validate repo contents.

## Engineering Handoff

### What Changed

- Added `.cairn/repos.yaml` as the attached-repo metadata file.
- Added `.cairn-workspace` as an explicit repo-local pointer back to the pod Cairn workspace.
- Added `cairn repo attach --name NAME --path RELPATH [--url URL] [--no-pointer]`.
- Added `cairn repo list`.
- Added `cairn repo discover [--from DIR]`.
- Discovery verifies that the resolved workspace has `.cairn/config.yaml`.
- CLI output explicitly says attached repos are reference metadata only and Cairn does not clone, index, sync, or validate repo contents.

### Validation Evidence

- `go test ./...` passed.
- Manual smoke with sibling layout `/cairn-kb`, `/repo-a`, and `/repo-b` passed:
  - initialized the Cairn workspace
  - attached both repos
  - listed both repos
  - discovered the workspace from `repo-a`

### Action And Approval Notes

- Action model: local write.
- No risky, sensitive, destructive, or production actions were taken.
- No human approval was required beyond the user's request to continue.

### Risks And Assumptions

- Repo paths are required to be relative to the Cairn workspace.
- Pointer writes are enabled by default for deterministic discovery, with `--no-pointer` available when the operator wants metadata only.
- This feature intentionally does not inspect or manage source repo contents.

### QA Focus Areas

- Verify attach records metadata and writes the pointer file.
- Verify list output includes repo names, paths, URLs, and the non-ownership boundary.
- Verify discover resolves from an attached repo through `.cairn-workspace`.
- Verify the fixture with one Cairn workspace and two sibling repos.

## QA Verdict

### Verdict

Pass.

### Evidence Summary

- `go test ./...` passed.
- Manual QA fixture with one Cairn workspace and two sibling repos passed.
- QA fixture verified `repo attach`, `repo list`, and `repo discover`.
- CLI output preserved the non-ownership boundary: Cairn will not clone, index, sync, or validate repo contents.

### Defects

None filed.

### Evidence Quality

Sufficient. Automated tests cover metadata recording, pointer writing/discovery, sorted loading, missing pointer errors, and CLI attach/list/discover behavior.

### State Transition

Moved to `done`.
