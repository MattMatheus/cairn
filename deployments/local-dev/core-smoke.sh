#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cairn() {
  (cd "$REPO_ROOT" && GOCACHE="${GOCACHE:-$REPO_ROOT/.cache/go-build}" go run ./cmd/cairn "$@")
}

require_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf 'Core smoke failed: %s\nExpected to find: %s\nOutput:\n%s\n' "$label" "$needle" "$haystack" >&2
    exit 1
  fi
}

write_local_fs_config() {
  local root="$1"
  local remote_root="$2"
  mkdir -p "$root/.cairn"
  cat >"$root/.cairn/config.yaml" <<YAML
schema_version: 1
workspace_id: cairn:workspace:core-smoke
managed_folders:
  - inbox
  - agents
  - working
  - decisions
  - runbooks
  - projects
  - services
  - handoffs
  - onboarding
  - archive
document_types:
  note: inbox
  handoff: handoffs
  investigation: working
  decision: decisions
  runbook: runbooks
  project: projects
  service: services
  onboarding: onboarding
remote_sync:
  provider: local_fs
  root: $remote_root
required_skills: []
YAML
}

WORK_ROOT="${WORK_ROOT:-$(mktemp -d "${TMPDIR:-/tmp}/cairn-core-smoke.XXXXXX")}"
REMOTE_ROOT="$WORK_ROOT/remote-store"
PRIMARY="$WORK_ROOT/primary"
PULLED="$WORK_ROOT/pulled"
CONFLICT="$WORK_ROOT/conflict"
PHRASE="cairn-core-smoke-$(date +%s)"

printf 'Initializing primary workspace: %s\n' "$PRIMARY"
cairn --root "$PRIMARY" init --workspace-id cairn:workspace:core-smoke >/dev/null
write_local_fs_config "$PRIMARY" "$REMOTE_ROOT"

printf 'Capturing and validating local document...\n'
cairn --root "$PRIMARY" capture \
  --actor codex \
  --title CoreSmokeSeed \
  --type note \
  --tags core,smoke \
  --body "The Cairn Core local filesystem smoke contains $PHRASE." >/dev/null

validate_out="$(cairn --root "$PRIMARY" validate)"
require_contains "$validate_out" "Workspace validation passed" "validate"

printf 'Refreshing local index and searching locally...\n'
refresh_out="$(cairn --root "$PRIMARY" index refresh)"
require_contains "$refresh_out" "Local index refreshed: true" "local index refresh"
search_out="$(cairn --root "$PRIMARY" search --mode auto --query "$PHRASE")"
require_contains "$search_out" "Found 1 result(s)." "local search count"
require_contains "$search_out" "agents/codex/coresmokeseed.md" "local search path"

printf 'Pushing to local filesystem remote store...\n'
status_out="$(cairn --root "$PRIMARY" sync status)"
require_contains "$status_out" "Sync diverged: false" "initial sync status"
push_out="$(cairn --root "$PRIMARY" sync push)"
require_contains "$push_out" "Sync push applied: true" "initial sync push"

printf 'Creating a second workspace from the synced base...\n'
cp -R "$PRIMARY" "$PULLED"

printf 'Creating and pushing a remote-side document...\n'
cairn --root "$PRIMARY" capture \
  --actor codex \
  --title CoreSmokeRemote \
  --type note \
  --tags core,smoke \
  --body "Remote create for $PHRASE." >/dev/null
push_out="$(cairn --root "$PRIMARY" sync push)"
require_contains "$push_out" "Sync push applied: true" "remote create push"

printf 'Pulling into second workspace and searching locally there...\n'
pull_out="$(cairn --root "$PULLED" sync pull)"
require_contains "$pull_out" "Sync pull applied: true" "sync pull"
test -f "$PULLED/agents/codex/coresmokeremote.md"
refresh_out="$(cairn --root "$PULLED" index refresh)"
require_contains "$refresh_out" "Local index refreshed: true" "pulled local index refresh"
search_out="$(cairn --root "$PULLED" search --mode auto --query "$PHRASE")"
require_contains "$search_out" "Found 2 result(s)." "pulled local search count"

printf 'Preparing deterministic divergence check...\n'
cairn --root "$PRIMARY" capture \
  --actor codex \
  --title CoreSmokeConflict \
  --type note \
  --tags core,smoke \
  --body "Conflict base for $PHRASE." >/dev/null
push_out="$(cairn --root "$PRIMARY" sync push)"
require_contains "$push_out" "Sync push applied: true" "conflict base push"

cp -R "$PRIMARY" "$CONFLICT"
printf '\nRemote divergent edit for %s.\n' "$PHRASE" >>"$PRIMARY/agents/codex/coresmokeconflict.md"
push_out="$(cairn --root "$PRIMARY" sync push)"
require_contains "$push_out" "Sync push applied: true" "remote divergent push"

printf '\nLocal divergent edit for %s.\n' "$PHRASE" >>"$CONFLICT/agents/codex/coresmokeconflict.md"
conflict_out="$(cairn --root "$CONFLICT" sync status)"
require_contains "$conflict_out" "Sync diverged: true" "conflict status divergence"
require_contains "$conflict_out" "Conflict:" "conflict status details"

printf '\nCore smoke passed.\n'
printf 'Workspace root: %s\n' "$WORK_ROOT"
printf 'Remote store: %s\n' "$REMOTE_ROOT"
printf 'Search phrase: %s\n' "$PHRASE"
