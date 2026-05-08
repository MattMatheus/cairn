# Planning Note: Code Review Repairs (findings-5-8)

- `date`: 2026-05-08
- `source`: human request following code review (`findings-5-8.md`)
- `next_stage`: pm

## Goal
Convert the 2026-05-08 code review findings (1 Critical, 7 High, 11 Medium, 8 Low) into bounded, testable engineering stories and route them through the harness.

## Constraints
- Pilot is in flight. Repairs must not regress `make pilot-check`.
- Default branch is `dev`; cycle commits land via `cycle-{cycle_id}` per `flywheel.yaml`.
- Flywheel/operational files are out of scope for code changes.
- Terraform changes (H5) require explicit approval before apply; story scope ends at PR-ready Terraform.

## Assumptions
- Critical and High findings are in scope this cycle. Medium/Low are bundled into one quality refinement story to keep cycle bounded.
- The indexer container source is out-of-tree; H5 mitigations are limited to ingress-side controls in `deployments/terraform`.
- No need to change the `Flywheel/` directory.

## Risks
- YAML-escaping fix (H6) could break existing stored frontmatter on round-trip if not handled carefully — must include round-trip tests.
- Filtering `.cairn/config.yaml` from sync (H3) changes manifest shape; existing remote stores may carry stale entries — pull-side tolerance required.
- ACA ingress IP restriction (H5) could lock out legitimate clients during deploy if not staged.

## Scope Boundary
- In: code repairs to `internal/*`, `scripts/install.sh`, `deployments/terraform/`, plus tests.
- Out: indexer container source (separate repo), VS Code extension publish, release packaging.

## Created Intake Artifacts
1. `STORY-20260508-install-checksum-fail-closed.md` — Critical (C1)
2. `STORY-20260508-sync-phantom-conflict.md` — High (H1)
3. `STORY-20260508-mcp-large-request-and-yaml-escape.md` — High (H2, H6)
4. `STORY-20260508-path-traversal-hardening.md` — High (H4, H7)
5. `STORY-20260508-workspace-config-not-shared.md` — High (H3)
6. `STORY-20260508-aca-ingress-hardening.md` — High (H5)
7. `STORY-20260508-quality-refinement-bundle.md` — Medium bundle (M1-M11) + selected Low

## Next Stage Recommendation
- `pm`: rank the seven stories, place the top item in `active`, leave the rest in `ready`.
