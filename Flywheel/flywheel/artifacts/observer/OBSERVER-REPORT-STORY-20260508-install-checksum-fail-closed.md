# Observer Report: STORY-20260508-install-checksum-fail-closed

## Metadata
- `cycle_id`: STORY-20260508-install-checksum-fail-closed
- `generated_at_utc`: 2026-05-08T15:42:58Z
- `branch`: dev
- `story_path`: Flywheel/flywheel/backlog/engineering/done/STORY-20260508-install-checksum-fail-closed.md
- `actor`: matt.matheus@healthcatalyst.com

## Diff Inventory
- A	findings-5-8.md
- A	Flywheel/flywheel/artifacts/planning/PLAN-20260508-code-review-repairs.md
- A	Flywheel/flywheel/backlog/engineering/done/STORY-20260508-install-checksum-fail-closed.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-aca-ingress-hardening.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-mcp-large-request-and-yaml-escape.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-path-traversal-hardening.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-quality-refinement-bundle.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-sync-phantom-conflict.md
- A	Flywheel/flywheel/backlog/engineering/ready/STORY-20260508-workspace-config-not-shared.md
- M	Flywheel/flywheel/backlog/engineering/active/README.md
- M	Flywheel/flywheel/backlog/engineering/ready/README.md
- M	scripts/install.sh

## Objective
- `intended_outcome`: Make `scripts/install.sh` fail closed when the release `.sha256` cannot be downloaded, eliminating the supply-chain risk where curl-piped install accepted unverified binaries (Critical, C1 from `findings-5-8.md`).
- `scope_boundary`: Single-script change. No `install.ps1` edits (already fail-closed). No release-pipeline changes.

## Inputs And Evidence
- `artifacts_reviewed`: `findings-5-8.md`, `scripts/install.sh`, `scripts/install.ps1`, `PLAN-20260508-code-review-repairs.md`.
- `tools_used`: Edit, Bash (`sh -n`, isolation reproduction script, `make pilot-check`).
- `external_sources`: none.

## Changes Made
- `files_changed`:
  - `scripts/install.sh` — checksum download failure now fatal; verification path otherwise unchanged.
  - `findings-5-8.md` — code review report added at repo root.
  - `Flywheel/flywheel/artifacts/planning/PLAN-20260508-code-review-repairs.md` — planning note for the repair cycle.
  - Seven engineering stories under `Flywheel/flywheel/backlog/engineering/{ready,done}/`.
  - `Flywheel/flywheel/backlog/engineering/{active,ready}/README.md` — updated queues.
- `state_transitions`: `STORY-20260508-install-checksum-fail-closed`: intake → active → qa → done.
- `non_file_actions`: ran `flywheel/tools/validate_intake_items.sh` (PASS); ran `make pilot-check` (PASS).

## Validation
- `checks_run`: `sh -n scripts/install.sh`; isolation reproduction (`/tmp/test-checksum-fail.sh` against `https://example.invalid/...`); `make pilot-check`.
- `results`: syntax clean; isolation script exited 1 with the new error message and never reached install; pilot-check passed end-to-end.
- `checks_not_run`: live happy-path verification against a real published release tag (network/release dependency); deferred to next release cut.

## Workflow Sync Checks
- [x] Entry docs updated if workflow behavior changed. (n/a; no workflow contract change)
- [x] Prompts updated if stage behavior changed. (n/a)
- [x] Process docs updated if contracts or gates changed. (n/a)
- [x] Queue order and state remain synchronized.

## Warnings And Risks
- `unresolved_risks`: AC2 (valid checksum → unchanged install) validated structurally only, not against a live release.
- `assumptions_carried`: release pipeline always publishes a `.sha256` next to the binary asset.
- `warnings`: six remaining repair stories sit in `ready` and must be sequenced (path-traversal → workspace-config → phantom-conflict → MCP+YAML → ACA ingress → quality bundle). ACA ingress story requires explicit human approval before any `terraform apply`.

## Action Record
- `highest_action_class`: local edit; no remote effects, no installs performed, no Terraform applied.
- `approval_required`: no.
- `approval_reference`: n/a.

## Next Step
- `recommended_next_state`: pm — activate `STORY-20260508-path-traversal-hardening` next.
- `follow_up_work`: six remaining repair stories listed in `ready/README.md`.
- `durable_promotions`: cycle commit `cycle-STORY-20260508-install-checksum-fail-closed` on branch `dev`.

## Release Impact
- Release scope: `required` for the next release; users on the curl-pipe install path immediately benefit.
- Additional release actions: ensure release workflow continues to publish `.sha256` per asset; consider documenting the change in release notes under "security".
