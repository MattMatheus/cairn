# Observer Report: Remote Profile Config Client Wiring

## Result
- Accepted.

## Work Completed
- Added remote sync and remote index config fields.
- Wired Azure Blob remote store construction from workspace config.
- Wired remote index HTTP client construction from workspace config.
- Added Azure CLI token boundary for indexer audience tokens.
- Preserved local-only behavior when remote config is absent.

## Verification
- `GOCACHE=/private/tmp/cairn-go-cache go test ./...`

## QA Notes
- Workspace config stores endpoint/account/container/audience metadata only.
- No client secrets, storage account keys, or long-lived tokens are required.
- Token acquisition remains lazy through Azure CLI provider boundaries.

## Next Suggested Step
- Promote `STORY-20260504-cli-purge-archived-document`.
