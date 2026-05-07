package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/document"
)

func TestSetupLocalSyncInitializesWorkspaceAndConfiguresRemote(t *testing.T) {
	root := t.TempDir()
	remoteRoot := filepath.Join(t.TempDir(), "remote")

	result, err := SetupLocalSync(root, SetupLocalSyncOptions{
		WorkspaceID: "cairn:workspace:pilot",
		RemoteRoot:  remoteRoot,
	})
	if err != nil {
		t.Fatalf("SetupLocalSync() error = %v", err)
	}
	if result.WorkspaceID != "cairn:workspace:pilot" || result.ConfigPath != ".cairn/config.yaml" {
		t.Fatalf("unexpected setup result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, ".cairn", "config.yaml")); err != nil {
		t.Fatalf("config.yaml was not created: %v", err)
	}
	cfg, err := document.LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "local_fs" || cfg.RemoteSync.Root != remoteRoot {
		t.Fatalf("unexpected remote sync config: %#v", cfg.RemoteSync)
	}
}

func TestSetupLocalSyncIsIdempotentForSameRemote(t *testing.T) {
	root := t.TempDir()
	remoteRoot := filepath.Join(t.TempDir(), "remote")
	if _, err := SetupLocalSync(root, SetupLocalSyncOptions{RemoteRoot: remoteRoot}); err != nil {
		t.Fatalf("first SetupLocalSync() error = %v", err)
	}
	if _, err := SetupLocalSync(root, SetupLocalSyncOptions{RemoteRoot: remoteRoot}); err != nil {
		t.Fatalf("second SetupLocalSync() should be idempotent: %v", err)
	}
}

func TestSetupLocalSyncRefusesDifferentExistingRemoteSync(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
remote_sync:
  provider: azure_blob
  account: acct
  container: cairn
`)

	_, err := SetupLocalSync(root, SetupLocalSyncOptions{RemoteRoot: "/tmp/cairn-remote"})
	if err == nil || !strings.Contains(err.Error(), "remote_sync is already configured") {
		t.Fatalf("expected existing remote_sync error, got %v", err)
	}
}

func TestSetupLocalSyncRequiresRemoteRoot(t *testing.T) {
	if _, err := SetupLocalSync(t.TempDir(), SetupLocalSyncOptions{}); err == nil {
		t.Fatalf("expected remote root requirement")
	}
}

func TestSetupAzureSyncInitializesWorkspaceAndConfiguresRemote(t *testing.T) {
	root := t.TempDir()

	result, err := SetupAzureSync(root, SetupAzureSyncOptions{
		WorkspaceID: "cairn:workspace:pilot",
		Account:     "cairnpilot",
		Container:   "pod-a",
	})
	if err != nil {
		t.Fatalf("SetupAzureSync() error = %v", err)
	}
	if result.WorkspaceID != "cairn:workspace:pilot" || result.ConfigPath != ".cairn/config.yaml" {
		t.Fatalf("unexpected setup result: %#v", result)
	}
	cfg, err := document.LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.RemoteSync.Provider != "azure_blob" || cfg.RemoteSync.Account != "cairnpilot" || cfg.RemoteSync.Container != "pod-a" || cfg.RemoteSync.Prefix != "" {
		t.Fatalf("unexpected remote sync config: %#v", cfg.RemoteSync)
	}
}

func TestSetupAzureSyncRequiresContainer(t *testing.T) {
	if _, err := SetupAzureSync(t.TempDir(), SetupAzureSyncOptions{Account: "cairnpilot"}); err == nil {
		t.Fatalf("expected container requirement")
	}
}
