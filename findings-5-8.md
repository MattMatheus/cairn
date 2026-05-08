# Cairn Code Review — 2026-05-08

## Summary

Reviewed: `cmd/cairn`, `internal/*` (10 packages), `extensions/vscode-cairn`, `scripts/`, `deployments/`, top-level Go module files (~15k lines of Go + JS + shell + Terraform). `Flywheel/` excluded per request.

**Counts:** 1 Critical · 7 High · 11 Medium · 8 Low

**Top 5 fixes to land first:**

1. `scripts/install.sh:60-74` — checksum fetch failure proceeds without integrity verification (curl-pipe install). `install.ps1` correctly fails closed; shell script must match.
2. `internal/syncstate/manifest.go:60-95` + `internal/workspace/init.go:197-203` — sync push uploads `.cairn/config.yaml` and `.cairn/repos.yaml`, propagating one user's local Azure credentials/paths to all collaborators.
3. `internal/syncstate/compare.go:168-170` — synthesizes a phantom `Conflict` pairing two unrelated changes when both sides have non-overlapping edits, then refuses sync.
4. `internal/mcpserver/server.go:99-122` — `bufio.Scanner` default 64KB buffer silently drops MCP requests with large `body` arguments (e.g. capture_note with long markdown).
5. `internal/workspace/repos.go:232-238` — `cleanRepoPath` accepts `..` and parent-relative paths; `repo attach --path ..` writes `.cairn-workspace` outside the workspace tree.

---

## Critical

### C1. install.sh accepts unverified binaries when checksum fetch fails
- **File:** `scripts/install.sh:60-74`
- **What:** `if curl -fsSL "$checksum_url" -o ... ; then verify ; else echo Warning ; fi` — if the `.sha256` URL 404s or the network fails, install proceeds with no integrity check.
- **Why:** README invites `curl ... | sh`. An attacker who can serve releases (or MITM unprotected mirrors) but not the checksum file gets unverified execution. `install.ps1` correctly fails closed (line 53-54); the shell script must match.
- **Fix:** `exit 1` on checksum download failure, or require `--insecure` to opt out.

---

## High

### H1. Sync-comparison synthesizes a phantom conflict on unrelated changes
- **File:** `internal/syncstate/compare.go:168-170`
- **What:** When `Diverged` is true (both sides have changes) but no path/document overlaps, the code appends `Conflict{Local: localChanges[0], Remote: remoteChanges[0]}` — pairing two unrelated changes as if they were the same conflict.
- **Why:** Surfaces a conflict that does not exist; `PlanFromStatus` (`plan.go:26-32`) refuses sync.
- **Fix:** When `len(result) == 0` but diverged, leave conflicts empty and let the caller render "diverged with no overlap" separately.

### H2. MCP server input scanner truncates at 64 KB
- **File:** `internal/mcpserver/server.go:99-122`
- **What:** `bufio.NewScanner(input)` uses the default `MaxScanTokenSize = 65536`. MCP `tools/call` requests embedding a `body` argument exceeding 64 KB cause `scanner.Scan()` to return `false` and the server exits silently with `bufio.ErrTooLong`.
- **Why:** Silent drop of long requests; user sees no response.
- **Fix:** `scanner.Buffer(make([]byte, 0, 1<<20), 16<<20)` or use `json.NewDecoder(input)`.

### H3. Default `.cairnignore` does not exclude `.cairn/config.yaml` or `.cairn/repos.yaml`
- **File:** `internal/workspace/init.go:197-203`; manifest at `internal/syncstate/manifest.go:60-95`
- **What:** `.cairn/sync-state.json` and `.cairn/remote-manifest.json` are filtered programmatically (`status.go:91-102`), but `.cairn/config.yaml`, `.cairn/repos.yaml`, and `.cairn/schemas/` are pushed to the shared remote store.
- **Why:** A single team member's local Azure account/endpoint or attached-repo paths get propagated to all collaborators. Sync becomes "last writer wins" for workspace settings.
- **Fix:** Extend `statusComparableEntries` to filter `.cairn/config.yaml` and `.cairn/repos.yaml`, or update default `.cairnignore`.

### H4. `cleanRepoPath` accepts `..` and parent-relative paths
- **File:** `internal/workspace/repos.go:232-238`
- **What:** Only rejects `.` and empty. Inputs like `..`, `../sibling`, or `../../etc` survive. `AttachRepo` then `Stat`s the resolved absolute path and (with `WritePointer` default true) writes `.cairn-workspace` into that arbitrary directory (`repos.go:104`).
- **Why:** Pointer file written outside the workspace tree.
- **Fix:** Reject results that are `..`, start with `../`, or are absolute. Match `internal/document/lifecycle.go:484-493` (`cleanWorkspacePath`).

### H5. ACA indexer is `external_enabled = true` with no IP allowlist
- **File:** `deployments/terraform/main.tf:101-166` (ingress 118-127)
- **What:** Container App ingress is public, no `ip_security_restriction` blocks. Client (`internal/remoteindex/http.go:67-75`) sends a Bearer token only if configured; omits header when token is empty (line 72).
- **Why:** If the indexer container does not reject anonymous requests, anyone on the internet can call `/index/refresh` and `/search`. The indexer source isn't in this repo, so the threat model rests on assumptions about the deployed container.
- **Fix:** Add `ip_security_restriction` rules; require client cert / EasyAuth; or front with API Management / App Gateway + WAF. Document required indexer-side authentication.

### H6. Frontmatter rendering does not YAML-escape strings
- **File:** `internal/document/lifecycle.go:406-431` (`renderDocument`, `writeStringArray`)
- **What:** `fmt.Sprintf("title: %s\n", metadata.Title)` writes raw values. A captured title containing `:`, leading `-`, `#`, `[`, `&`, or a newline produces invalid YAML.
- **Why:** Capture is invoked from `cli.runNote` and `mcpops.CaptureNote`, both accepting arbitrary input. Round-trip parse may yield wrong values silently.
- **Fix:** Quote strings with YAML-special characters (mirror `quoteRepoValue` in `repos.go:247`).

### H7. `pull.go` workspace-path traversal check is incomplete on Windows
- **File:** `internal/syncstate/pull.go:228-234`
- **What:** `clean[:3] == "../"` only catches POSIX-style escapes. On Windows `filepath.Clean("../foo")` returns `..\foo`, slipping past the check.
- **Why:** Maliciously-crafted remote manifest entries with `..\..\Windows\System32\foo` resolve outside the workspace on Windows hosts.
- **Fix:** `strings.HasPrefix(clean, ".."+string(filepath.Separator))` or `filepath.Rel(root, joined)` and reject `..` components — matches `remotestore/local_fs.go:130-147`.

---

## Medium

### M1. Hardcoded Azurite shared key
- **File:** `internal/remotestore/azure_blob.go:252` — `defaultKey = "Y2Fpcm4tbG9jYWwtZGV2LWF6dXJpdGUta2V5LTAwMDE="`
- **What:** Embedded base64 key for `azurite` auth mode. Auth-mode is gated to localhost endpoints only (line 65-67), so impact is limited; the key still ships in the binary.
- **Fix:** Require `AZURITE_ACCOUNT_KEY` env var explicitly, or use the well-known Azurite default with an in-source comment.

### M2. Variable shadowing in `Search`
- **File:** `internal/localindex/search.go:43-57`
- **What:** Inner `results, err := i.Query(...)` shadows the outer `var results []mcpschema.SearchResult` (line 41). Semantics are correct today; invites future bugs.
- **Fix:** Use `=` or rename the inner variable.

### M3. UTF-8 unsafe slicing in snippets and summaries
- **Files:** `internal/localindex/search.go:269-279` (`snippet`); `internal/mcpops/read.go:210-216` (`summarize`)
- **What:** `content[start:end]` slices on byte boundaries; `strings.Index` against `strings.ToLower` of UTF-8 may not align with the original byte offsets. Output may contain mojibake or invalid UTF-8 in JSON envelopes.
- **Fix:** Operate on `[]rune` or back up to a `utf8.RuneStart` boundary.

### M4. LIKE-wildcard handling in metadata queries
- **File:** `internal/localindex/index.go:197-228`
- **What:** Tag/text/actor filters concatenate into LIKE patterns: `"%" + query.Text + "%"` and `"%\""+query.Tag+"\"%"`. Parameters are bound (no SQL injection), but `%`/`_` in user input give unintended matches; tag values containing `"` corrupt the JSON-equality contract.
- **Fix:** Escape `%`, `_`, `\` in bound values; add `ESCAPE '\'` to LIKE clauses. For tag exact-match, use a join table.

### M5. `unquoteConfig` strips ALL leading/trailing quote characters
- **File:** `internal/document/config.go:614-619`
- **What:** `strings.Trim(value, "\"")` is a cutset, not a delimiter strip. `"hello"world"` becomes `hello"world`. Same bug in `internal/workspace/repos.go:251-256`.
- **Fix:** Match leading/trailing quote pairs explicitly (mirror `internal/document/frontmatter.go:204-212` `unquote`).

### M6. `LoadConfig` is called per-document during validation/health
- **Files:** `internal/workspace/validate.go:213-233`; `internal/workspace/health.go:85` (via `isManagedMarkdown`)
- **What:** `isManagedMarkdown` re-reads `.cairn/config.yaml` for every markdown file walked.
- **Fix:** Load config once per call site and pass through.

### M7. `repairMetadata` doesn't normalize bad-but-non-zero values
- **File:** `internal/document/lifecycle.go:296-335`
- **What:** Only fills defaults when fields are zero/empty. Created in the future, empty entries inside non-empty actor lists, or `created > updated` are not normalized; durable-mode `Validate` at line 142-147 then blocks promotion with a confusing error.
- **Fix:** Add normalization for in-list empties and obviously bogus timestamps.

### M8. Dead/redundant block in promote path
- **File:** `internal/document/lifecycle.go:142-161`
- **What:** Lines 158-161 set `metadata.Status = "proposed"; metadata.Updated = w.now()` after the canonical/proposed branch already ran. Conceptually dead — `repairMetadata` should own status assignment.
- **Fix:** Move status assignment into a single branch.

### M9. `health.go` does not honor `ctx` between phases
- **File:** `internal/workspace/health.go:43-133`
- **What:** Only the per-file loop checks `ctx.Err()`. Phase boundaries (`Validate`, `syncstate.StatusReport`) don't check; large workspaces respond slowly to Ctrl-C.
- **Fix:** Add `if err := ctx.Err(); err != nil { return ..., err }` before each phase.

### M10. `nextADRNumber` race on concurrent promotions
- **File:** `internal/document/lifecycle.go:355-382`
- **What:** Read `decisions/`, compute next number, write file. Two concurrent `promote --status canonical` (CLI + MCP, or two MCP servers) can collide on `ADR-0042-foo.md`.
- **Fix:** Open destination with `O_EXCL`; increment on `EEXIST`.

### M11. `runSearch` always re-indexes the workspace
- **File:** `internal/cli/cli.go:1056-1097` (line 1072)
- **What:** Every `cairn search` invocation calls `local.Index.IndexWorkspace`. For large workspaces this is slow and conflicts with `cairn index refresh`.
- **Fix:** Only re-index if the index is missing/empty.

---

## Low

### L1. `splitCSV` does not handle quoted commas
- **File:** `internal/cli/cli.go:1199-1208` — `--tags 'a, "b, c"'` becomes three tags.

### L2. Friendly sync error matches raw substring
- **File:** `internal/cli/cli.go:1328-1336` — `strings.Contains(err.Error(), "remote store is required")` couples CLI to error text in `syncstate/{push,pull}.go`. Use a typed error or sentinel.

### L3. `parseFrontmatterLines` mixes loop-counter idioms
- **File:** `internal/document/frontmatter.go:73-117` — `for i := 0; ...; i++` plus manual `i++` in the body. Works but fragile.

### L4. `usage()` text is one giant line
- **File:** `internal/cli/cli.go:1216-1219` — wrap with newlines per command.

### L5. ADR id format inconsistency
- **File:** `internal/document/lifecycle.go:364` — regex `\d{4,}` allows 5+ digits; format string `%04d` always pads to 4. Post-9999 produces 5-digit IDs.

### L6. `health.go` walks the workspace twice
- **File:** `internal/workspace/health.go:71-103` then `Validate(...)` (line 116) walks again via `markdownPaths`.

### L7. `slugifyActor` can produce empty actor
- **File:** `internal/cli/cli.go:878-894` — usernames starting with `--` slug to empty; capture then errors with "actor is required".

### L8. Unused import in extension
- **File:** `extensions/vscode-cairn/src/extension.js:2` — `const path = require("path")` not used.

---

## Notes / non-issues

- `internal/remotestore/local_fs.go:130-147` correctly uses `filepath.Rel` for path traversal — good model.
- `internal/syncstate/pull.go:62-99` does pre-fetch + backup + transactional rollback — well-designed.
- `internal/document/lifecycle.go:484-493` `cleanWorkspacePath` is the right pattern; repo-attach equivalent should match.
- `extensions/vscode-cairn/src/cairnCli.js` correctly uses `execFile` (no shell) and `workspaceRelativePath` rejects `..` — solid.
