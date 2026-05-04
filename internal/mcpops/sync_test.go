package mcpops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/mcpschema"
	"cairn/internal/remoteindex"
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

func TestSyncPullSuggestsRemoteIndexRefreshWhenConfigured(t *testing.T) {
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

	local := &Local{Root: root, RemoteStore: store, RemoteIndex: &remoteindex.FakeClient{}}
	envelope, err := local.SyncPull(context.Background(), mcpschema.SyncRequest{})
	if err != nil {
		t.Fatalf("SyncPull() error = %v", err)
	}
	found := false
	for _, step := range envelope.NextSteps {
		if step.Action == string(mcpschema.ToolIndexRefresh) {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected index refresh next step: %#v", envelope.NextSteps)
	}
}

func TestSyncPushAppliesLocalOnlyPlan(t *testing.T) {
	root := t.TempDir()
	base := syncstate.Manifest{ManifestVersion: syncstate.ManifestVersion, WorkspaceID: "pod-1"}
	if err := syncstate.Save(root, syncstate.State{Entries: base.Entries}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	writeSyncRemoteManifest(t, root, base)
	writeFile(t, root, "working/local.md", managedMarkdown("cairn:local", "Local", "local", "note", "working", nil, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Local body."))
	store := remotestore.NewMemoryStore()

	local := &Local{Root: root, RemoteStore: store}
	envelope, err := local.SyncPush(context.Background(), mcpschema.SyncRequest{})
	if err != nil {
		t.Fatalf("SyncPush() error = %v", err)
	}
	if !envelope.OK || !envelope.Data.Applied || len(envelope.Data.ChangedPaths) != 1 {
		t.Fatalf("unexpected push envelope: %#v", envelope)
	}
	if _, ok, err := store.ReadObject(context.Background(), "working/local.md"); err != nil || !ok {
		t.Fatalf("expected pushed remote object ok=%t err=%v", ok, err)
	}
	if _, ok, err := store.ReadManifest(context.Background()); err != nil || !ok {
		t.Fatalf("expected pushed remote manifest ok=%t err=%v", ok, err)
	}
}

func TestSyncPushRefusesWhenLiveRemoteDivergesFromLocalFixture(t *testing.T) {
	root := t.TempDir()
	base := syncstate.Manifest{ManifestVersion: syncstate.ManifestVersion, WorkspaceID: "pod-1"}
	if err := syncstate.Save(root, syncstate.State{Entries: base.Entries}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	writeSyncRemoteManifest(t, root, base)
	writeFile(t, root, "working/local.md", managedMarkdown("cairn:local", "Local", "local", "note", "working", nil, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Local body."))

	remoteContent := managedMarkdown("cairn:remote", "Remote", "remote", "note", "working", nil, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Remote body.")
	liveRemote := syncstate.Manifest{
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
	store := &trackingStore{MemoryStore: remotestore.NewMemoryStore()}
	if err := store.WriteObject(context.Background(), "working/remote.md", []byte(remoteContent)); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
	}
	if err := store.WriteManifest(context.Background(), liveRemote); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	store.operations = nil

	local := &Local{Root: root, RemoteStore: store}
	envelope, err := local.SyncPush(context.Background(), mcpschema.SyncRequest{})
	if err == nil {
		t.Fatal("expected SyncPush to refuse divergent live remote manifest")
	}
	if envelope.OK || envelope.Data.Applied {
		t.Fatalf("expected refused envelope, got %#v", envelope)
	}
	assertOperationOrder(t, store.operations, "read-manifest", "write-object")
	assertNoOperation(t, store.operations, "write-manifest")
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

type trackingStore struct {
	*remotestore.MemoryStore
	operations []string
}

func (s *trackingStore) ReadManifest(ctx context.Context) (syncstate.Manifest, bool, error) {
	s.operations = append(s.operations, "read-manifest")
	return s.MemoryStore.ReadManifest(ctx)
}

func (s *trackingStore) WriteManifest(ctx context.Context, manifest syncstate.Manifest) error {
	s.operations = append(s.operations, "write-manifest")
	return s.MemoryStore.WriteManifest(ctx, manifest)
}

func (s *trackingStore) WriteObject(ctx context.Context, path string, content []byte) error {
	s.operations = append(s.operations, "write-object:"+path)
	return s.MemoryStore.WriteObject(ctx, path, content)
}

func (s *trackingStore) DeleteObject(ctx context.Context, path string) error {
	s.operations = append(s.operations, "delete-object:"+path)
	return s.MemoryStore.DeleteObject(ctx, path)
}

func assertOperationOrder(t *testing.T, operations []string, before string, forbiddenPrefix string) {
	t.Helper()
	beforeIndex := -1
	for i, operation := range operations {
		if operation == before && beforeIndex == -1 {
			beforeIndex = i
		}
		if strings.HasPrefix(operation, forbiddenPrefix) && (beforeIndex == -1 || i < beforeIndex) {
			t.Fatalf("operation %q occurred before %q: %#v", operation, before, operations)
		}
	}
	if beforeIndex == -1 {
		t.Fatalf("missing operation %q in %#v", before, operations)
	}
}

func assertNoOperation(t *testing.T, operations []string, forbidden string) {
	t.Helper()
	for _, operation := range operations {
		if operation == forbidden || strings.HasPrefix(operation, forbidden+":") {
			t.Fatalf("unexpected operation %q in %#v", forbidden, operations)
		}
	}
}

func testHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
