# Cycle Navigation

## Flow
1. Confirm the current branch matches `workflow.required_branch`.
2. Run the required stage.
3. Read the selected prompt from `paths.prompts`.
4. Work only in configured backlog and artifact locations.
5. Move work through explicit queue states.
6. Close the cycle with an observer report and one cycle commit.

## Empty Queue Rule
If the configured engineering active queue is empty, report `no stories` and route work to planning or PM refinement.

