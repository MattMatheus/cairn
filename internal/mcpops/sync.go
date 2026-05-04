package mcpops

import (
	"context"

	"cairn/internal/mcpschema"
	"cairn/internal/syncstate"
)

func (l *Local) SyncStatus(ctx context.Context, _ mcpschema.EmptyRequest) (mcpschema.Envelope[mcpschema.SyncStatusData], error) {
	status, err := syncstate.StatusReport(ctx, l.Root, syncstate.StatusOptions{
		WorkspaceID: l.provenance("sync_status").WorkspaceID,
		Now:         l.Now,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncStatusData]{}, err
	}
	envelope := syncstate.Envelope(status)
	envelope.Provenance = l.provenance("sync_status")
	return envelope, nil
}

func (l *Local) SyncDryRun(ctx context.Context, _ mcpschema.SyncRequest) (mcpschema.Envelope[mcpschema.SyncMutationData], error) {
	status, err := syncstate.StatusReport(ctx, l.Root, syncstate.StatusOptions{
		WorkspaceID: l.provenance("sync_dry_run").WorkspaceID,
		Now:         l.Now,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncMutationData]{}, err
	}
	envelope := syncstate.PlanEnvelope(status)
	envelope.Provenance = l.provenance("sync_dry_run")
	return envelope, nil
}

func (l *Local) SyncPull(ctx context.Context, _ mcpschema.SyncRequest) (mcpschema.Envelope[mcpschema.SyncMutationData], error) {
	status, err := syncstate.StatusReport(ctx, l.Root, syncstate.StatusOptions{
		WorkspaceID: l.provenance("sync_pull").WorkspaceID,
		Now:         l.Now,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncMutationData]{}, err
	}
	plan, err := syncstate.ApplyPull(ctx, l.Root, status, l.RemoteStore, syncstate.PullOptions{Now: l.Now})
	if err != nil {
		envelope := syncstate.PlanEnvelope(status)
		envelope.Provenance = l.provenance("sync_pull")
		envelope.Data.Applied = false
		return envelope, err
	}
	envelope := syncstate.PlanEnvelope(status)
	envelope.OK = true
	envelope.Data.Applied = true
	envelope.Data.ChangedPaths = changedPaths(plan.Changes)
	envelope.Provenance = l.provenance("sync_pull")
	envelope.NextSteps = []mcpschema.NextStep{{
		Action: string(mcpschema.ToolSyncStatus),
		Label:  "Run sync status",
		Reason: "Confirm the workspace is clean after pull.",
	}}
	return envelope, nil
}

func changedPaths(changes []syncstate.Change) []mcpschema.ChangedPath {
	out := make([]mcpschema.ChangedPath, 0, len(changes))
	for _, change := range changes {
		out = append(out, mcpschema.ChangedPath{Path: change.Path, Kind: string(change.Type)})
	}
	return out
}
