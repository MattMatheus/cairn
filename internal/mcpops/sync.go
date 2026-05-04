package mcpops

import (
	"context"

	"cairn/internal/mcpschema"
	"cairn/internal/syncstate"
)

func (l *Local) SyncStatus(ctx context.Context, _ mcpschema.EmptyRequest) (mcpschema.Envelope[mcpschema.SyncStatusData], error) {
	status, err := l.syncStatus(ctx, "sync_status")
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncStatusData]{}, err
	}
	envelope := syncstate.Envelope(status)
	envelope.Provenance = l.provenance("sync_status")
	return envelope, nil
}

func (l *Local) SyncDryRun(ctx context.Context, _ mcpschema.SyncRequest) (mcpschema.Envelope[mcpschema.SyncMutationData], error) {
	status, err := l.syncStatus(ctx, "sync_dry_run")
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncMutationData]{}, err
	}
	envelope := syncstate.PlanEnvelope(status)
	envelope.Provenance = l.provenance("sync_dry_run")
	return envelope, nil
}

func (l *Local) SyncPull(ctx context.Context, _ mcpschema.SyncRequest) (mcpschema.Envelope[mcpschema.SyncMutationData], error) {
	status, err := l.syncStatus(ctx, "sync_pull")
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
	if l.RemoteIndex != nil {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolIndexRefresh),
			Label:  "Refresh remote index",
			Reason: "Pulled workspace changes may need remote index refresh.",
		})
	}
	return envelope, nil
}

func (l *Local) SyncPush(ctx context.Context, _ mcpschema.SyncRequest) (mcpschema.Envelope[mcpschema.SyncMutationData], error) {
	status, err := l.syncStatus(ctx, "sync_push")
	if err != nil {
		return mcpschema.Envelope[mcpschema.SyncMutationData]{}, err
	}
	plan, err := syncstate.ApplyPush(ctx, l.Root, status, l.RemoteStore, syncstate.PushOptions{
		WorkspaceID: l.provenance("sync_push").WorkspaceID,
		Now:         l.Now,
	})
	if err != nil {
		envelope := syncstate.PlanEnvelope(status)
		envelope.Provenance = l.provenance("sync_push")
		envelope.Data.Applied = false
		return envelope, err
	}
	envelope := syncstate.PlanEnvelope(status)
	envelope.OK = true
	envelope.Data.Applied = true
	envelope.Data.ChangedPaths = changedPaths(plan.Changes)
	envelope.Provenance = l.provenance("sync_push")
	envelope.NextSteps = []mcpschema.NextStep{{
		Action: string(mcpschema.ToolSyncStatus),
		Label:  "Run sync status",
		Reason: "Confirm the workspace is clean after push.",
	}}
	return envelope, nil
}

func (l *Local) syncStatus(ctx context.Context, source string) (syncstate.Status, error) {
	opts := syncstate.StatusOptions{
		WorkspaceID: l.provenance(source).WorkspaceID,
		Now:         l.Now,
	}
	if l.RemoteStore != nil {
		manifest, ok, err := l.RemoteStore.ReadManifest(ctx)
		if err != nil {
			return syncstate.Status{}, err
		}
		if ok {
			opts.RemoteManifest = &manifest
		}
	}
	return syncstate.StatusReport(ctx, l.Root, opts)
}

func changedPaths(changes []syncstate.Change) []mcpschema.ChangedPath {
	out := make([]mcpschema.ChangedPath, 0, len(changes))
	for _, change := range changes {
		out = append(out, mcpschema.ChangedPath{Path: change.Path, Kind: string(change.Type)})
	}
	return out
}
