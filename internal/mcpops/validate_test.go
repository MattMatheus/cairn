package mcpops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/mcpschema"
)

func TestValidateWorkspaceReturnsEnvelope(t *testing.T) {
	root := t.TempDir()
	writeValidateFile(t, root, "working/missing.md", "# Missing frontmatter\n")
	local := &Local{Root: root}
	envelope, err := local.ValidateWorkspace(context.Background(), mcpschema.ValidateWorkspaceRequest{})
	if err != nil {
		t.Fatalf("ValidateWorkspace() error = %v", err)
	}
	if !envelope.OK {
		t.Fatalf("expected warning-only workspace validation to return OK")
	}
	if len(envelope.Data.Findings) == 0 {
		t.Fatalf("expected findings collection")
	}
	if len(envelope.NextSteps) == 0 {
		t.Fatalf("expected next step guidance")
	}
	if envelope.Provenance.Source != "workspace_validation" {
		t.Fatalf("unexpected provenance source %q", envelope.Provenance.Source)
	}
}

func TestValidateWorkspaceSurfacesConfigFindings(t *testing.T) {
	root := t.TempDir()
	writeValidateFile(t, root, ".cairn/config.yaml", `schema_version: nope
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
`)
	local := &Local{Root: root}
	envelope, err := local.ValidateWorkspace(context.Background(), mcpschema.ValidateWorkspaceRequest{})
	if err != nil {
		t.Fatalf("ValidateWorkspace() error = %v", err)
	}
	if envelope.OK {
		t.Fatalf("expected config error to make envelope not OK: %#v", envelope)
	}
	if len(envelope.Data.Findings) == 0 || envelope.Data.Findings[0].Path != ".cairn/config.yaml" {
		t.Fatalf("expected config finding in envelope: %#v", envelope)
	}
}

func writeValidateFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
