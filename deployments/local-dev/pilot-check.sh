#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
WORK_ROOT="${WORK_ROOT:-$(mktemp -d "${TMPDIR:-/tmp}/cairn-pilot-check.XXXXXX")}"
BIN_DIR="$WORK_ROOT/bin"
FRESH_ROOT="$WORK_ROOT/fresh-workspace"
FRESH_REMOTE_ROOT="$WORK_ROOT/fresh-remote-store"
SAMPLE_ROOT="$WORK_ROOT/sample-workspace"
GOCACHE="${GOCACHE:-/tmp/cairn-go-build-cache}"

require_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    printf 'Pilot check failed: %s\nExpected to find: %s\nOutput:\n%s\n' "$label" "$needle" "$haystack" >&2
    exit 1
  fi
}

printf 'Running Go test suite...\n'
(cd "$REPO_ROOT" && GOCACHE="$GOCACHE" go test ./...)

printf 'Building pilot binary...\n'
mkdir -p "$BIN_DIR"
(cd "$REPO_ROOT" && GOCACHE="$GOCACHE" go build -o "$BIN_DIR/cairn" ./cmd/cairn)

printf 'Checking help output...\n'
help_out="$("$BIN_DIR/cairn" help)"
require_contains "$help_out" "usage: cairn" "help"

printf 'Checking version output...\n'
version_out="$("$BIN_DIR/cairn" version)"
require_contains "$version_out" "cairn dev" "version"

printf 'Checking fresh workspace init and local sync setup...\n'
init_out="$("$BIN_DIR/cairn" --root "$FRESH_ROOT" init --workspace-id cairn:workspace:pilot-check)"
require_contains "$init_out" "Initialized workspace cairn:workspace:pilot-check" "fresh init"
test -f "$FRESH_ROOT/.cairn/config.yaml"
doctor_out="$("$BIN_DIR/cairn" --root "$FRESH_ROOT" doctor)"
require_contains "$doctor_out" "Config: present" "fresh doctor config"
setup_out="$("$BIN_DIR/cairn" --root "$FRESH_ROOT" setup local-sync --remote-root "$FRESH_REMOTE_ROOT")"
require_contains "$setup_out" "Configured local sync in .cairn/config.yaml" "fresh local sync setup"
require_contains "$(cat "$FRESH_ROOT/.cairn/config.yaml")" "provider: local_fs" "fresh local sync provider"
doctor_out="$("$BIN_DIR/cairn" --root "$FRESH_ROOT" doctor --remote)"
require_contains "$doctor_out" "Remote check: reachable" "fresh doctor remote"

printf 'Validating sample workspace fixture...\n'
cp -R "$REPO_ROOT/examples/pilot-workspace" "$SAMPLE_ROOT"
validate_out="$("$BIN_DIR/cairn" --root "$SAMPLE_ROOT" validate)"
require_contains "$validate_out" "Workspace validation passed." "sample validate"

printf 'Refreshing and searching sample workspace...\n'
refresh_out="$("$BIN_DIR/cairn" --root "$SAMPLE_ROOT" index refresh)"
require_contains "$refresh_out" "Local index refreshed: true" "sample index refresh"
search_out="$("$BIN_DIR/cairn" --root "$SAMPLE_ROOT" search --query pilot-handshake)"
require_contains "$search_out" "Found 1 result(s)." "sample search count"
require_contains "$search_out" "runbooks/pilot-handshake.md" "sample search path"

printf 'Running no-service local sync smoke...\n'
(cd "$REPO_ROOT" && GOCACHE="$GOCACHE" deployments/local-dev/core-smoke.sh)

printf '\nPilot check passed.\n'
printf 'Work root: %s\n' "$WORK_ROOT"
