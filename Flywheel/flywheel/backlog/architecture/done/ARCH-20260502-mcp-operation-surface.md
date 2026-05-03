# Architecture Story: MCP Operation Surface ADR

## Metadata
- `id`: ARCH-20260502-mcp-operation-surface
- `owner_role`: Software Architect
- `status`: done
- `source`: direct
- `decision_refs`: [ADR-mcp-operation-surface]
- `decision_owner`: Software Architect
- `success_metric`: A reviewer can implement the v1 MCP server surface without exposing raw filesystem primitives or unsafe delete behavior.

## Decision Scope
- Define Cairn's v1 MCP operation surface, including tool names, request/response schemas at decision level, actor identity, profile use, validation behavior, sync/index tool boundaries, and progressive read/search behavior.

## Problem Statement
- Agents need a stable MCP surface for finding, writing, promoting, archiving, syncing, and reading Cairn knowledge. The tool set must expose product operations, not arbitrary filesystem access, while giving agents enough metadata and next-step guidance to work safely.

## Inputs
- Existing decisions:
  - `docs/product/north-star.md`
- Existing architecture artifacts:
  - None yet.
- Constraints:
  - Initial tools include `get_bootstrap`, `capture_note`, `promote_document`, `archive_document`, `read_document`, `find_document`, `search_context`, `list_documents`, `validate_workspace`, `sync_status`, `sync_pull`, `sync_push`, `index_status`, and `index_refresh`.
  - MCP tools may archive and move documents, but must not expose hard purge/delete in v1.
  - MCP tools should use product operations rather than raw filesystem primitives.
  - `read_document` supports `summary`, `frontmatter`, `toc`, `sections`, and `full`.
  - `search_context` supports `auto`, `metadata`, `full_text`, and `semantic`.
  - `auto` search reports attempted modes, unavailable modes, warnings, and suggested next steps.

## Outputs Required
- Decision updates:
  - `docs/adr/ADR-mcp-operation-surface.md` defining v1 MCP tools, operation boundaries, input/output schema shape, error/warning shape, actor/profile handling, and safety restrictions.
- Architecture artifacts:
  - Tool matrix with purpose, mutability, profile needs, and validation behavior.
  - Common response envelope for warnings, next steps, and provenance.
  - Read/search result shape.
  - Permission boundary for archive versus purge.
- Risks and tradeoffs:
  - Too many tools can create confusing agent behavior.
  - Too few tools can push agents back to direct filesystem edits.
  - Large `read_document` responses can burn context.
  - Remote profile tools may fail when Azure auth or sync is unavailable.

## Alternatives Considered
- Expose raw filesystem read/write tools.
- Combine all document writes into one generic mutation tool.
- Keep sync and index operations CLI-only.
- Return full document contents by default.

## Operational Impact
- Sets agent ergonomics and safety boundaries.
- Determines how external agents discover bootstrap context and leave durable notes.
- Shapes CLI/MCP parity expectations.

## Acceptance Criteria
1. ADR defines the v1 MCP tool list and each tool's intended operation.
2. ADR distinguishes read-only, local write, remote/sync, and forbidden operations.
3. ADR defines common response metadata, warnings, provenance, and suggested next steps.
4. ADR defines read/search progressive disclosure behavior.
5. ADR confirms hard delete/purge is not exposed through MCP in v1.

## Review Focus
- Confirm agents can complete expected workflows without direct filesystem primitives.
- Confirm the surface is small enough to support well.
- Confirm profile and auth assumptions are visible in tool behavior.

## Next Step
- Architecture QA should review `docs/adr/ADR-mcp-operation-surface.md`.

## Architecture Handoff
- `Architecture decision`: Drafted `docs/adr/ADR-mcp-operation-surface.md` as a proposed ADR covering the v1 MCP tool matrix, safety boundaries, common response shape, actor/profile handling, and read/search progressive disclosure.
- `Alternatives considered`: Raw filesystem tools, one generic mutation tool, CLI-only sync/index operations, and full reads by default.
- `Key risks`: Tool surface can grow confusing; too little structure pushes agents to direct file edits; remote-profile tools need graceful failure.
- `Follow-on implementation paths`: Concrete JSON schemas, common response envelope, lifecycle-aware mutations, profile-aware sync/index tools, and tests for forbidden purge/delete exposure.
- `Next state recommendation`: Move to architecture QA.

## QA Review
- `Verdict`: Pass.
- `Evidence summary`: ADR was reviewed against the north-star document and the story acceptance criteria. It defines the v1 MCP tool list, operation mutability, common response shape, actor/profile behavior, read/search progressive disclosure, and excludes hard delete/purge.
- `Evidence quality call`: Sufficient for architecture acceptance; concrete per-tool JSON schemas are properly left as follow-on work.
- `Defects`: None.
- `Required fixes`: None.
- `Split decision`: No split required now. Concrete MCP schemas can be implementation stories unless tool contracts become large enough to require child ADRs.
- `Completed work summary`: Accepted the MCP operation surface ADR.
- `Next suggested or required step`: Use this ADR to seed implementation stories for MCP schema definitions, shared response envelopes, lifecycle-aware mutations, and purge/delete exposure tests.
- `Next state recommendation`: Move to architecture done.

## Intake Promotion Checklist
- [x] Decision scope is explicit and bounded.
- [x] Problem statement explains why the decision is needed now.
- [x] Inputs are listed and available.
- [x] Outputs are concrete and reviewable.
- [x] Alternatives and operational impact are explicit.
- [x] Follow-on implementation work is split out when needed.
