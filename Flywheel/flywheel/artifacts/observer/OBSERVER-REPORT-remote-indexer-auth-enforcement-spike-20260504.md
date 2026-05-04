# Observer Report: Remote Indexer Auth Enforcement Spike

## Summary
- Documented the V1 remote indexer auth enforcement recommendation.
- Recommended ACA built-in Microsoft Entra authentication at the edge.
- Kept workspace/pod authorization inside the indexer using authenticated principal claims.
- Identified custom JWT middleware as a fallback only after tenant validation.

## QA
- Reviewed against north-star auth constraints.
- Ran `GOCACHE=/private/tmp/cairn-go-cache go test ./...`.

## Next Suggested Step
- Plan the follow-up implementation story for ACA auth configuration and indexer authorization shim.
