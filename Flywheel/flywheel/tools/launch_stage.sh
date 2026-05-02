#!/usr/bin/env bash
set -euo pipefail

stage="${1:-engineering}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/lib/config.sh"

root_dir="$(flywheel_repo_root)"
required_branch="$(flywheel_config_get workflow.required_branch)"
prompts_dir="$(flywheel_path paths.prompts)"
eng_active_dir="$(flywheel_path paths.engineering.active)"
arch_active_dir="$(flywheel_path paths.architecture.active)"

print_artifact_workflow() {
  local stage_name="$1"
  local artifact_output

  if ! flywheel_feature_enabled "integrations.artifact_workflow.enabled"; then
    return 0
  fi

  artifact_output="$("$script_dir/artifact_workflow.sh" "$stage_name")"
  if [[ -n "$artifact_output" ]]; then
    printf '%s\n' "$artifact_output"
  fi
}

if ! git -C "$root_dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "abort: not a git repository at $root_dir" >&2
  exit 1
fi

current_branch="$(git -C "$root_dir" branch --show-current)"
if [[ "$current_branch" != "$required_branch" ]]; then
  echo "abort: active branch is '$current_branch'; expected '$required_branch'" >&2
  exit 1
fi

select_top_item() {
  local lane_dir="$1"
  local lane_readme="$lane_dir/README.md"
  local candidate

  if [[ -f "$lane_readme" ]]; then
    while IFS= read -r candidate; do
      [[ -z "$candidate" ]] && continue
      if [[ "$candidate" != /* ]]; then
        candidate="$lane_dir/$candidate"
      fi
      if [[ -f "$candidate" ]]; then
        echo "$candidate"
        return 0
      fi
    done < <(sed -En 's/^[[:space:]]*[0-9]+\.[[:space:]]*`([^`]+)`.*/\1/p' "$lane_readme")
  fi

  find "$lane_dir" -maxdepth 1 -type f -name '*.md' ! -name 'README.md' | sort | head -n1
}

case "$stage" in
  planning)
    cat <<EOF
launch: ${prompts_dir}/planning.md
cycle: planning
checklist:
  1) capture goals, constraints, risks, and assumptions
  2) create or refine intake artifacts
  3) recommend next stage
  4) do not implement production changes
EOF
    print_artifact_workflow planning
    ;;
  architect)
    top_item="$(select_top_item "$arch_active_dir" || true)"
    if [[ -z "${top_item:-}" ]]; then
      echo "no stories"
      exit 0
    fi
    rel_item="${top_item#"$root_dir"/}"
    cat <<EOF
launch: ${prompts_dir}/architect.md
cycle: architect
story: ${rel_item}
checklist:
  1) restate decision scope
  2) update architecture artifacts
  3) prepare handoff
  4) move work to architecture qa
EOF
    print_artifact_workflow architect
    ;;
  engineering)
    top_item="$(select_top_item "$eng_active_dir" || true)"
    if [[ -z "${top_item:-}" ]]; then
      echo "no stories"
      exit 0
    fi
    rel_item="${top_item#"$root_dir"/}"
    cat <<EOF
launch: ${prompts_dir}/engineering.md
cycle: engineering
story: ${rel_item}
checklist:
  1) implement the selected story
  2) update validation
  3) prepare handoff
  4) move work to engineering qa
  5) do not commit yet
EOF
    print_artifact_workflow engineering
    ;;
  qa)
    cat <<EOF
launch: ${prompts_dir}/qa.md
cycle: qa
checklist:
  1) validate the story in the engineering qa lane
  2) file bugs if needed
  3) move work to done or back to active
  4) run observer
  5) create the cycle commit
EOF
    print_artifact_workflow qa
    ;;
  pm)
    cat <<EOF
launch: ${prompts_dir}/pm.md
cycle: pm
checklist:
  1) refine intake work
  2) validate metadata and lane placement
  3) rank the active queues
  4) keep queue ordering explicit
EOF
    print_artifact_workflow pm
    ;;
  cycle)
    cat <<EOF
launch: ${prompts_dir}/cycle.md
cycle: engineering+qa loop
loop:
  - run engineering stage
  - if output is "no stories": stop
  - run qa stage
  - run observer
  - create one cycle commit
  - repeat
EOF
    print_artifact_workflow cycle
    ;;
  *)
    echo "usage: flywheel/tools/launch_stage.sh [planning|architect|engineering|qa|pm|cycle]" >&2
    exit 1
    ;;
esac
