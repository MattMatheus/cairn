# Architecture Story: Document Model And Lifecycle ADR

## Metadata
- `id`: ARCH-20260502-document-model-lifecycle
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-document-model-lifecycle]
- `decision_owner`: Software Architect
- `success_metric`: A reviewer can implement capture, validation, promotion, archive, and ADR numbering behavior without resolving lifecycle semantics ad hoc.

## Decision Scope
- Define Cairn's v1 managed document model, core frontmatter, document identity, type/status rules, promotion lifecycle, archive/purge boundary, and ADR numbering behavior.

## Problem Statement
- The north-star establishes portable markdown as the source of truth while requiring reliable canonical knowledge for agents. V1 needs a precise lifecycle contract before implementation so capture, validation, promotion, sync, indexing, and MCP tools do not each invent different rules.

## Inputs
- Existing decisions:
  - `docs/product/north-star.md`
- Existing architecture artifacts:
  - None yet.
- Constraints:
  - Markdown remains readable without Cairn.
  - Cairn-managed canonical/synced/indexed documents require valid core frontmatter.
  - Unknown frontmatter fields warn, not fail.
  - ADR numbers are assigned only when a `decision` becomes `canonical`.
  - Promotion transforms one existing document rather than creating original-versus-promoted copies.
  - Purge is CLI-only; agents may archive but not hard delete.

## Outputs Required
- Decision updates:
  - `docs/adr/ADR-document-model-lifecycle.md` defining document identity, required core frontmatter, filename/slug rules, status transitions, document type mapping, promotion semantics, archive layout, purge boundary, and ADR numbering.
- Architecture artifacts:
  - Example frontmatter block.
  - State transition table.
  - Promotion behavior table for `proposed` and `canonical`.
  - ADR filename/numbering rule.
- Risks and tradeoffs:
  - Too much strictness breaks portable markdown use.
  - Too much permissiveness weakens agent reliability.
  - Late ADR number assignment requires careful concurrency handling.

## Alternatives Considered
- Require valid frontmatter for every markdown file before any Cairn operation.
- Keep discovery permissive but block only canonical promotion, sync, and indexing.
- Assign ADR numbers at draft creation.
- Copy promoted documents instead of transforming them.

## Operational Impact
- Establishes the contract for validation, sync eligibility, search/index eligibility, and MCP write operations.
- Determines how teams repair bypassed or hand-written markdown.

## Acceptance Criteria
1. ADR states exactly when missing/invalid frontmatter is a warning versus a blocking error.
2. ADR defines all v1 statuses and allowed transitions.
3. ADR defines promotion to `proposed` versus `canonical`.
4. ADR defines decision ADR numbering and final path assignment.
5. ADR defines archive and purge permissions for CLI and MCP.

## Review Focus
- Confirm the lifecycle is strict enough for agents but still preserves portable markdown.
- Confirm no v2 privacy/retention requirements are pulled into v1.

## Next Step
- Architecture QA should review `docs/adr/ADR-document-model-lifecycle.md`.

## Architecture Handoff
- `Architecture decision`: Drafted `docs/adr/ADR-document-model-lifecycle.md` as a proposed ADR covering document identity, frontmatter, lifecycle status, promotion, archive/purge, and ADR numbering.
- `Alternatives considered`: Strict frontmatter for all operations, permissive discovery with durable-boundary blocking, draft-time ADR numbering, and copy-on-promotion.
- `Key risks`: ADR number allocation may need concurrency handling with sync; validation must avoid making portable markdown feel brittle.
- `Follow-on implementation paths`: Frontmatter parsing/validation, id generation, capture/promotion, archive/purge, ADR number allocation, lifecycle tests.
- `Next state recommendation`: Move to architecture QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: ADR was reviewed against the north-star document and the story acceptance criteria. It covers warning versus blocking frontmatter rules, v1 statuses and transitions, proposed versus canonical promotion, ADR numbering, archive, and CLI-only purge.
- `Evidence quality call`: Sufficient for architecture acceptance; implementation details remain correctly deferred to follow-on work.
- `Defects`: None.
- `Required fixes`: None.
- `Split decision`: No split required now. ADR number allocation/concurrency can be handled as follow-on implementation detail or sync refinement if needed.
- `Completed work summary`: Accepted the document model and lifecycle ADR.
- `Next suggested or required step`: Use this ADR to seed implementation stories for frontmatter validation, capture, promotion, archive, purge, and ADR numbering.
- `Next state recommendation`: Move to architecture done.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
