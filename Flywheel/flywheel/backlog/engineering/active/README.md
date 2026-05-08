# Engineering Active Queue

Ordered execution queue for engineering implementation.

## Active Sequence

No active engineering stories. (`STORY-20260508-install-checksum-fail-closed` moved to QA on 2026-05-08.)

## Next Ready Candidates

1. `STORY-20260508-path-traversal-hardening` — High (H4, H7). Security hardening for `cleanRepoPath` and Windows-aware pull containment.
2. `STORY-20260508-workspace-config-not-shared` — High (H3). Stop leaking local Azure/repo config via sync.
3. `STORY-20260508-sync-phantom-conflict` — High (H1). Eliminate fake conflict on diverged-but-disjoint state.
4. `STORY-20260508-mcp-large-request-and-yaml-escape` — High (H2, H6). MCP buffer + frontmatter YAML escaping.
5. `STORY-20260508-aca-ingress-hardening` — High (H5). Terraform ingress IP restriction; needs human approval before `terraform apply`.
6. `STORY-20260508-quality-refinement-bundle` — Medium + Low bundle. Land last to avoid merge churn.

## Batch Goal

Polish Cairn for internal pilot rollout by reducing first-run friction, supporting non-AI developers, and keeping one pod knowledge base usable alongside multiple code repositories.

## PM Notes

- Planning note: `Flywheel/flywheel/artifacts/planning/PLAN-20260508-pilot-polish-roadmap.md`.
- PM activated the first pilot polish story on 2026-05-08.
- `STORY-20260508-doctor-full-pilot-readiness` moved to QA on 2026-05-08.
- PM activated `STORY-20260508-interactive-capture-flow` on 2026-05-08.
- `STORY-20260508-interactive-capture-flow` moved to QA on 2026-05-08.
- PM activated `STORY-20260508-repo-attachment-discovery` on 2026-05-08.
- `STORY-20260508-repo-attachment-discovery` moved to QA on 2026-05-08.
- PM activated `STORY-20260508-ado-lifecycle-candidate-capture` on 2026-05-08.
- `STORY-20260508-ado-lifecycle-candidate-capture` moved to QA on 2026-05-08.
- PM activated `STORY-20260508-vscode-workspace-helpers` on 2026-05-08.
- `STORY-20260508-vscode-workspace-helpers` moved to QA on 2026-05-08.
- PM activated `STORY-20260508-knowledge-health-report` on 2026-05-08.
- `STORY-20260508-knowledge-health-report` moved to QA on 2026-05-08.
- Follow-on intake order is interactive capture, repo attachment/discovery, ADO lifecycle candidate capture, VS Code helpers, and knowledge health reporting.
- 2026-05-08: PM ranked seven repair stories from `findings-5-8.md` (planning note `PLAN-20260508-code-review-repairs.md`). Activated `STORY-20260508-install-checksum-fail-closed` first; remaining six in ranked ready order. ACA ingress story requires human approval before any `terraform apply`.
