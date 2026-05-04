# Observer Report: Index Refresh Contract Wiring

## Result
- Accepted.

## Work Completed
- Added remote index refresh operation adapter.
- Added refresh response fields for accepted, refreshed, job id, last refresh time, and message.
- Preserved graceful degradation for missing and failing remote indexers.
- Added CLI support for `cairn index refresh`.
- Added post-pull next-step suggestion to refresh remote index when configured.

## Verification
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Notes
- Accepted asynchronous refresh does not imply synchronous completion.
- Refreshed responses suggest searching refreshed context.
- Missing remote indexer reports warning, unavailable mode, and next step.
- Failed remote refresh reports retryable degradation.

## Next Suggested Step
- Promote `STORY-20260503-mcp-mutating-tools-gated`.
