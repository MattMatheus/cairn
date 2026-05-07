package document

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadConfigUsesRemoteProfileSettings(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: azure_blob
  account: acct
  container: cairn
  prefix: pod-a
remote_index:
  url: https://indexer.example
  audience: api://cairn-indexer
  tenant_id: tenant
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "azure_blob" || cfg.RemoteSync.Account != "acct" || cfg.RemoteSync.Container != "cairn" || cfg.RemoteSync.Prefix != "pod-a" {
		t.Fatalf("unexpected remote sync config: %#v", cfg.RemoteSync)
	}
	if cfg.RemoteIndex.URL != "https://indexer.example" || cfg.RemoteIndex.Audience != "api://cairn-indexer" || cfg.RemoteIndex.TenantID != "tenant" {
		t.Fatalf("unexpected remote index config: %#v", cfg.RemoteIndex)
	}
}

func TestLoadConfigUsesLocalFSAndAzuriteSettings(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: local_fs
  root: /tmp/cairn-local-remote
  prefix: pod-a
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "local_fs" || cfg.RemoteSync.Root != "/tmp/cairn-local-remote" || cfg.RemoteSync.Prefix != "pod-a" {
		t.Fatalf("unexpected local_fs config: %#v", cfg.RemoteSync)
	}

	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: azure_blob
  endpoint: http://localhost:10000/devstoreaccount1
  container: cairn
  auth_mode: azurite
`)
	cfg, err = LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "azure_blob" || cfg.RemoteSync.Endpoint != "http://localhost:10000/devstoreaccount1" || cfg.RemoteSync.AuthMode != "azurite" {
		t.Fatalf("unexpected azurite config: %#v", cfg.RemoteSync)
	}
}

func TestLoadConfigUsesGeneratedPodRemoteProfileSettings(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
profiles:
  local:
    enabled: true
  pod-remote:
    enabled: true
    provider: azure_blob
    account: acct
    container: cairn
    prefix: pod-a
    url: https://indexer.example
    audience: api://cairn-indexer
    tenant_id: tenant
`)

	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "azure_blob" || cfg.RemoteSync.Account != "acct" || cfg.RemoteSync.Container != "cairn" || cfg.RemoteSync.Prefix != "pod-a" {
		t.Fatalf("unexpected remote sync config: %#v", cfg.RemoteSync)
	}
	if cfg.RemoteIndex.URL != "https://indexer.example" || cfg.RemoteIndex.Audience != "api://cairn-indexer" || cfg.RemoteIndex.TenantID != "tenant" {
		t.Fatalf("unexpected remote index config: %#v", cfg.RemoteIndex)
	}
}

func TestValidateConfigFilesReportsEnabledPodRemoteMissingRequiredFields(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
profiles:
  pod-remote:
    enabled: true
    provider: azure_blob
`)

	findings := ValidateConfigFiles(root)
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "pod-remote account or endpoint is required when enabled")
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "pod-remote container is required when enabled")
}

func TestValidateConfigFilesAcceptsLocalFSPodRemote(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
profiles:
  pod-remote:
    enabled: true
    provider: local_fs
    root: .cairn/local-remote
`)

	findings := ValidateConfigFiles(root)
	for _, finding := range findings {
		if finding.Severity == "error" {
			t.Fatalf("unexpected error finding: %#v", findings)
		}
	}
}

func TestValidateConfigFilesReportsMalformedConfig(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: two
workspace_id:
managed_folders:
  not-a-list
document_types:
  runbook: ../outside
broken line
`)

	findings := ValidateConfigFiles(root)
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "schema_version must be 1")
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "workspace_id is required")
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "managed_folders entries must use list syntax")
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "document type destination is invalid")
	assertConfigFinding(t, findings, ".cairn/config.yaml", "error", "malformed top-level config entry")
}

func TestValidateConfigFilesWarnsForUnknownDocumentTypeMapping(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  memo: working
`)

	findings := ValidateConfigFiles(root)
	assertConfigFinding(t, findings, ".cairn/config.yaml", "warning", "unknown document type mapping memo")
}

func TestValidateConfigFilesReportsSchemaMissingCoreFields(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
`)
	writeSchema(t, root, "custom.yaml", `schema_version: 1
name: custom
required_fields:
  - id
  - title
`)

	findings := ValidateConfigFiles(root)
	assertConfigFinding(t, findings, ".cairn/schemas/custom.yaml", "error", "required_fields must include Cairn core field schema_version")
	assertConfigFinding(t, findings, ".cairn/schemas/custom.yaml", "error", "required_fields must include Cairn core field tags")
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

func writeSchema(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, ".cairn", "schemas", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func assertConfigFinding(t *testing.T, findings []ConfigFinding, path string, severity string, contains string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Path == path && finding.Severity == severity && strings.Contains(finding.Message, contains) {
			return
		}
	}
	t.Fatalf("missing config finding path=%s severity=%s contains=%q in %#v", path, severity, contains, findings)
}
