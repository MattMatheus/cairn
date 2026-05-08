# Story: ACA indexer ingress hardening

## Metadata
- `id`: STORY-20260508-aca-ingress-hardening
- `owner_role`: SRE
- `status`: ready
- `source`: planning
- `decision_refs`: [ARCH-20260507-core-v1-indexing-boundary-refresh]
- `success_metric`: ACA indexer ingress is not anonymous-internet-reachable; access requires either IP allowlist or platform-level auth.
- `release_scope`: required

## Problem Statement
- `deployments/terraform/main.tf:101-166` deploys the indexer Container App with `external_enabled = true` and no IP restrictions. The CLI client (`internal/remoteindex/http.go:67-75`) sends a Bearer token only when configured — the header is omitted entirely if the token is empty. Indexer-side enforcement is out of this repo, so ingress posture is the only verifiable control here (High, H5).

## Scope
- In:
  - Add an `ip_security_restriction` block (or equivalent) gated by a Terraform variable (`allowed_cidrs`).
  - Document required indexer-side authentication in `deployments/terraform/README.md`.
  - Stage rollout note (apply during low-traffic window; verify clients on allowlist).
- Out:
  - App Gateway / API Management front door.
  - Indexer container source changes.

## Assumptions
- Pilot is small enough that an IP allowlist is operationally acceptable for now.

## Acceptance Criteria
1. `terraform plan` against current state reflects ingress IP restriction with a sensible deny-default.
2. Variable `allowed_cidrs` is required (no default that opens the world).
3. README documents how to run the apply safely and how to add a CIDR.
4. Sample CIDR list shows pilot pattern (e.g. office VPN, GitHub-hosted runners).

## Validation
- Required checks:
  - `terraform fmt` and `terraform validate` clean.
  - `terraform plan` review shows expected changes.
- Additional checks:
  - Manual: dry-run plan against a scratch state file.

## Dependencies
- Operator must provide pilot CIDR list.

## Risks
- Misconfigured CIDRs lock out the pilot user; mitigate with rollback steps in the README.

## Open Questions
- Should we move to EasyAuth (AAD) instead of IP allowlist? (Backlog architecture story if pursued.)

## Next Step
- PM ranks; engineering implements; human approval required before any `terraform apply`.
