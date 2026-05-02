#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/lib/config.sh"

eng_intake="$(flywheel_path paths.engineering.intake)"
arch_intake="$(flywheel_path paths.architecture.intake)"
story_template="$(flywheel_template_path story)"
bug_template="$(flywheel_template_path bug)"
arch_template="$(flywheel_template_path architecture_story)"

failures=0

fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}

require_line() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! grep -Eq -- "$pattern" "$file"; then
    fail "$message ($file)"
  fi
}

for file in "$eng_intake"/*.md; do
  [[ -e "$file" ]] || continue
  [[ "$(basename "$file")" == "README.md" ]] && continue
  [[ "$file" == "$story_template" || "$file" == "$bug_template" ]] && continue
  base="$(basename "$file")"
  if [[ "$base" == STORY-* ]]; then
    require_line "$file" '^## Metadata$' 'missing metadata header'
    require_line "$file" '^[[:space:]]*-[[:space:]]*`id`:[[:space:]]*STORY-' 'story id must start with STORY-'
    require_line "$file" '^[[:space:]]*-[[:space:]]*`status`:[[:space:]]*(intake|ready|active|qa|done|blocked|archive)$' 'story status invalid'
  elif [[ "$base" == BUG-* ]]; then
    require_line "$file" '^## Metadata$' 'missing metadata header'
    require_line "$file" '^[[:space:]]*-[[:space:]]*`id`:[[:space:]]*BUG-' 'bug id must start with BUG-'
    require_line "$file" '^[[:space:]]*-[[:space:]]*`priority`:[[:space:]]*(P0|P1|P2|P3)$' 'bug priority invalid'
  else
    fail "unexpected file type in engineering intake: $file"
  fi
done

for file in "$arch_intake"/*.md; do
  [[ -e "$file" ]] || continue
  [[ "$(basename "$file")" == "README.md" ]] && continue
  [[ "$file" == "$arch_template" ]] && continue
  base="$(basename "$file")"
  if [[ "$base" != ARCH-* ]]; then
    fail "unexpected file type in architecture intake: $file"
    continue
  fi
  require_line "$file" '^## Metadata$' 'missing metadata header'
  require_line "$file" '^[[:space:]]*-[[:space:]]*`id`:[[:space:]]*ARCH-' 'architecture story id must start with ARCH-'
  require_line "$file" '^[[:space:]]*-[[:space:]]*`status`:[[:space:]]*(intake|ready|active|qa|done|blocked|archive)$' 'architecture status invalid'
done

if [[ "$failures" -ne 0 ]]; then
  exit 1
fi

echo "PASS: intake metadata and lane separation validation"
