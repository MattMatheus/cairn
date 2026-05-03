package mcpops

import (
	"context"
	"testing"

	"cairn/internal/mcpschema"
)

func TestValidateWorkspaceReturnsEnvelope(t *testing.T) {
	local := &Local{Root: t.TempDir()}
	envelope, err := local.ValidateWorkspace(context.Background(), mcpschema.ValidateWorkspaceRequest{})
	if err != nil {
		t.Fatalf("ValidateWorkspace() error = %v", err)
	}
	if !envelope.OK {
		t.Fatalf("expected warning-only workspace validation to return OK")
	}
	if envelope.Data.Findings == nil {
		t.Fatalf("expected findings collection")
	}
	if len(envelope.NextSteps) == 0 {
		t.Fatalf("expected next step guidance")
	}
	if envelope.Provenance.Source != "workspace_validation" {
		t.Fatalf("unexpected provenance source %q", envelope.Provenance.Source)
	}
}
