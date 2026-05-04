package mcpops

import (
	"context"
	"testing"

	"cairn/internal/mcpschema"
)

func TestSyncStatusReturnsEnvelope(t *testing.T) {
	local := &Local{Root: t.TempDir()}
	envelope, err := local.SyncStatus(context.Background(), mcpschema.EmptyRequest{})
	if err != nil {
		t.Fatalf("SyncStatus() error = %v", err)
	}
	if !envelope.OK {
		t.Fatalf("missing remote fixture alone should not fail status: %#v", envelope)
	}
	if len(envelope.Warnings) == 0 {
		t.Fatalf("expected missing remote warning")
	}
	if envelope.Provenance.Source != "sync_status" {
		t.Fatalf("unexpected provenance source %q", envelope.Provenance.Source)
	}
}

func TestSyncDryRunReturnsPlan(t *testing.T) {
	local := &Local{Root: t.TempDir()}
	envelope, err := local.SyncDryRun(context.Background(), mcpschema.SyncRequest{DryRun: true})
	if err != nil {
		t.Fatalf("SyncDryRun() error = %v", err)
	}
	if envelope.Data.Plan == nil {
		t.Fatalf("expected dry-run plan")
	}
	if envelope.Data.Plan.Direction != mcpschema.SyncDirectionClean {
		t.Fatalf("unexpected direction %q", envelope.Data.Plan.Direction)
	}
}
