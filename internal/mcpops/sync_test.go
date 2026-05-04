package mcpops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cairn/internal/mcpschema"
	"cairn/internal/remotestore"
	"cairn/internal/syncstate"
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

func TestSyncPullAppliesRemoteOnlyPlan(t *testing.T) {
	root := t.TempDir()
	base := syncstate.Manifest{ManifestVersion: syncstate.ManifestVersion, WorkspaceID: "pod-1"}
	if err := syncstate.Save(root, syncstate.State{Entries: base.Entries}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	remoteContent := managedMarkdown("cairn:remote", "Remote", "remote", "note", "working", nil, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Remote body.")
	remote := syncstate.Manifest{
		ManifestVersion: syncstate.ManifestVersion,
		WorkspaceID:     "pod-1",
		Entries: []syncstate.Entry{{
			Path:       "working/remote.md",
			Kind:       "file",
			Size:       int64(len(remoteContent)),
			Hash:       testHash(remoteContent),
			DocumentID: "cairn:remote",
			Status:     "working",
			Type:       "note",
		}},
	}
	writeSyncRemoteManifest(t, root, remote)
	store := remotestore.NewMemoryStore()
	if err := store.WriteObject(context.Background(), "working/remote.md", []byte(remoteContent)); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
	}

	local := &Local{Root: root, RemoteStore: store}
	envelope, err := local.SyncPull(context.Background(), mcpschema.SyncRequest{})
	if err != nil {
		t.Fatalf("SyncPull() error = %v", err)
	}
	if !envelope.OK || !envelope.Data.Applied || len(envelope.Data.ChangedPaths) != 1 {
		t.Fatalf("unexpected pull envelope: %#v", envelope)
	}
	if _, err := os.Stat(filepath.Join(root, "working", "remote.md")); err != nil {
		t.Fatalf("expected pulled file: %v", err)
	}
}

func writeSyncRemoteManifest(t *testing.T, root string, manifest syncstate.Manifest) {
	t.Helper()
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	path := filepath.Join(root, ".cairn", "remote-manifest.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func testHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
