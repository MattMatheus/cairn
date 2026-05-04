package document

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFallsBackToDefaults(t *testing.T) {
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.DestinationFolder("runbook") != "runbooks" {
		t.Fatalf("default runbook destination = %q", cfg.DestinationFolder("runbook"))
	}
	if cfg.DestinationFolder("unknown") != "working" {
		t.Fatalf("unknown type destination = %q", cfg.DestinationFolder("unknown"))
	}
	if _, ok := cfg.ManagedFolderSet()["working"]; !ok {
		t.Fatalf("default managed folders missing working: %#v", cfg.ManagedFolders)
	}
}

func TestLoadConfigUsesCustomMappings(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
  - knowledge
document_types:
  note: inbox
  runbook: ops/runbooks
  decision: architecture/decisions
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.WorkspaceID != "cairn:workspace:test" {
		t.Fatalf("unexpected workspace id %q", cfg.WorkspaceID)
	}
	if cfg.DestinationFolder("runbook") != "ops/runbooks" {
		t.Fatalf("custom runbook destination = %q", cfg.DestinationFolder("runbook"))
	}
	if cfg.DestinationFolder("decision") != "architecture/decisions" {
		t.Fatalf("custom decision destination = %q", cfg.DestinationFolder("decision"))
	}
	if _, ok := cfg.ManagedFolderSet()["knowledge"]; !ok {
		t.Fatalf("custom managed folder missing: %#v", cfg.ManagedFolders)
	}
}

func writeConfig(t *testing.T, root string, content string) {
	t.Helper()
	path := filepath.Join(root, ".cairn", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
