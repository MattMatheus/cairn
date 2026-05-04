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
