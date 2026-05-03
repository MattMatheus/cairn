#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$script_dir/lib/config.sh"

config_root="$(flywheel_config_root)"
harness_dir="$config_root/flywheel"

check_file() {
  local path="$1"
  if [[ -f "$path" ]]; then
    echo "PASS: $path"
  else
    echo "FAIL: missing $path" >&2
    return 1
  fi
}

check_dir() {
  local path="$1"
  if [[ -d "$path" ]]; then
    echo "PASS: $path"
  else
    echo "FAIL: missing $path" >&2
    return 1
  fi
}

check_file "$(flywheel_config_file)"
check_file "$harness_dir/CONFIG_SCHEMA.md"
check_file "$harness_dir/README.md"
check_file "$harness_dir/HUMANS.md"
check_file "$harness_dir/AGENTS.md"
check_file "$harness_dir/DEVELOPMENT_CYCLE.md"

check_dir "$(flywheel_path paths.prompts)"
check_dir "$(flywheel_path paths.roles)"
check_dir "$(flywheel_path paths.process)"
check_dir "$(flywheel_path paths.templates)"
check_dir "$(flywheel_path paths.artifacts.planning)"
check_dir "$(flywheel_path paths.artifacts.observer)"
check_dir "$(flywheel_path paths.engineering.active)"
check_dir "$(flywheel_path paths.architecture.active)"

check_file "$(flywheel_template_path story)"
check_file "$(flywheel_template_path bug)"
check_file "$(flywheel_template_path architecture_story)"
check_file "$(flywheel_template_path observer_report)"

check_file "$harness_dir/tools/launch_stage.sh"
check_file "$harness_dir/tools/artifact_workflow.sh"
check_file "$harness_dir/tools/artifact_workflow_commands.sh"
check_file "$harness_dir/tools/run_observer_cycle.sh"
check_file "$harness_dir/tools/validate_intake_items.sh"

bash "$harness_dir/tools/validate_intake_items.sh"

echo "Result: PASS"
