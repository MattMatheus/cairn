# Planning: Pilot Polish Roadmap

## Objective

Prepare Cairn for polished internal pilot use before broader pod rollout. The next work should reduce first-run friction, support non-AI developers, and make one pod knowledge base usable alongside multiple code repositories without pulling code repos into Cairn ownership.

## Product Direction

Cairn remains one knowledge workspace per pod. A pod may own multiple code repositories, but the Cairn workspace is singular and syncs through its own blob-backed workspace state.

Preferred local layout:

```text
/pod-workspace
  /cairn-kb
  /repo-a
  /repo-b
  /repo-c
```

Code repositories may point to or attach to the pod Cairn workspace. They should not each contain duplicated Cairn knowledge stores.

## Prioritized Work

1. `STORY-20260508-doctor-full-pilot-readiness`
2. `STORY-20260508-interactive-capture-flow`
3. `STORY-20260508-repo-attachment-discovery`
4. `STORY-20260508-ado-lifecycle-candidate-capture`
5. `STORY-20260508-vscode-workspace-helpers`
6. `STORY-20260508-knowledge-health-report`

## Scope Boundaries

In scope:
- CLI-first pilot readiness checks.
- Low-friction capture for developers who do not use AI assistants.
- Repo-aware metadata and discovery for multi-repo pods.
- ADO lifecycle hooks that create candidate knowledge for human review.
- Traditional IDE support, starting with VS Code.
- Markdown or CLI knowledge health reporting.

Out of scope:
- CocoIndex as required Cairn Core infrastructure.
- Cross-pod search.
- Cloud drive sync.
- Product dashboards.
- AI IDE-specific integrations such as Cursor.
- Autonomous canonical promotion.
- PR validation as the primary remote-blob safety boundary.

## Key Decisions From Planning Discussion

- Validation safety should live primarily at Cairn mutation boundaries: capture, promote, sync, and remote-visible operations.
- ADO PR validation is only useful when the pipeline materializes the Cairn workspace or when Cairn docs live in repo branches. It should not be treated as the primary guard for blob-managed documents.
- ADO integration should focus first on lifecycle events that create candidate Cairn notes when work items, PRs, incidents, or releases transition state.
- VS Code support is valuable because it removes context switching for ordinary editor workflows, not because it targets AI IDE behavior.
- CocoIndex remains deferred until demonstrated pod needs exceed local SQLite metadata and full-text search.

## Assumptions

- Pilot teams value polish and trust more than a large feature surface.
- Most pod knowledge bases will stay small enough for local SQLite metadata and full-text search.
- A single pod may own multiple repositories with different branch and release rhythms.
- Non-AI developers should be first-class users of capture, validation, promotion, and search.
- ADO is the enterprise lifecycle system to integrate with first.

## Risks

- Adding repo awareness could blur ownership if Cairn starts indexing or managing code repositories. Keep the boundary explicit: attach and reference repos, do not ingest them by default.
- ADO lifecycle hooks could overproduce low-value candidate notes. Start with explicit, configurable state transitions and working/proposed outputs only.
- VS Code extension scope could expand quickly. Keep the first version to capture, search, promote, validate, doctor, and open workspace.
- Health reporting could drift toward dashboard work. Keep the first version as CLI and generated markdown.

## Success Signals

- A pilot engineer can run one command and understand whether the workspace is ready.
- A non-AI developer can capture and promote useful knowledge without memorizing frontmatter or long CLI syntax.
- A pod with multiple repos can attach or discover one Cairn workspace without duplicating knowledge.
- ADO transitions can produce candidate knowledge that still respects Cairn human promotion.
- The first VS Code helper removes common shell context switching without introducing a new product surface.

## Next Stage Recommendation

Next stage: `pm`.

Reason: the direction is already bounded enough for engineering stories. PM should rank the intake queue, promote the first story to ready/active, and keep the follow-on stories ordered for pilot polish.
