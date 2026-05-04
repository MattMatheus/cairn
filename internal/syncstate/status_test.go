package syncstate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStatusReportClean(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	if status.Comparison.Diverged {
		t.Fatalf("expected clean status, got divergence: %#v", status.Comparison)
	}
	if len(status.Comparison.LocalChanges) != 0 || len(status.Comparison.RemoteChanges) != 0 {
		t.Fatalf("expected no changes: %#v", status.Comparison)
	}
	envelope := Envelope(status)
	if !envelope.OK || envelope.Data.Diverged {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
}

func TestStatusReportClassifiesLocalOnlyChanges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "working/local.md", managedDocument("cairn:local", "working", "note", "Local"))

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	if status.Comparison.Diverged {
		t.Fatalf("expected local-only change to be safe: %#v", status.Comparison)
	}
	assertChange(t, status.Comparison.LocalChanges, ChangeCreate, "working/local.md")
	if len(status.Comparison.RemoteChanges) != 0 {
		t.Fatalf("expected no remote changes: %#v", status.Comparison.RemoteChanges)
	}
}

func TestStatusReportClassifiesRemoteOnlyChanges(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	remote := base
	remote.Entries = append(remote.Entries, Entry{Path: "working/remote.md", Kind: "file", Size: 1, Hash: "remote", DocumentID: "cairn:remote", Status: "working", Type: "note"})
	writeRemoteManifest(t, root, remote)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	if status.Comparison.Diverged {
		t.Fatalf("expected remote-only change to be safe: %#v", status.Comparison)
	}
	assertChange(t, status.Comparison.RemoteChanges, ChangeCreate, "working/remote.md")
	if len(status.Comparison.LocalChanges) != 0 {
		t.Fatalf("expected no local changes: %#v", status.Comparison.LocalChanges)
	}
}

func TestStatusReportDivergenceRefusesAndDoesNotMutate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	beforeState := readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))
	beforeFile := readBytes(t, filepath.Join(root, "working", "a.md"))

	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A Local"))
	remote := base
	for i := range remote.Entries {
		if remote.Entries[i].Path == "working/a.md" {
			remote.Entries[i].Hash = "remote"
			remote.Entries[i].Size++
		}
	}
	writeRemoteManifest(t, root, remote)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	if !status.Comparison.Diverged || len(status.Comparison.Conflicts) == 0 {
		t.Fatalf("expected diverged conflict: %#v", status.Comparison)
	}
	envelope := Envelope(status)
	if envelope.OK || len(envelope.NextSteps) == 0 {
		t.Fatalf("expected refusal envelope with next steps: %#v", envelope)
	}
	if string(beforeState) != string(readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))) {
		t.Fatalf("sync status mutated local state")
	}
	afterFile := readBytes(t, filepath.Join(root, "working", "a.md"))
	if string(beforeFile) == string(afterFile) {
		t.Fatalf("test setup failed to create local file change")
	}
	if !strings.Contains(string(afterFile), "A Local") {
		t.Fatalf("sync status mutated local workspace file:\n%s", string(afterFile))
	}
}

func TestStatusReportMissingRemoteManifestIsLocalOnly(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	if status.RemoteAvailable {
		t.Fatalf("expected missing remote fixture")
	}
	if status.Comparison.Diverged {
		t.Fatalf("missing remote fixture should not diverge: %#v", status.Comparison)
	}
	envelope := Envelope(status)
	if len(envelope.Warnings) == 0 {
		t.Fatalf("expected remote unavailable warning: %#v", envelope)
	}
}

func mustGenerate(t *testing.T, root string) Manifest {
	t.Helper()
	manifest, err := Generate(root, GenerateOptions{
		WorkspaceID: "pod-1",
		Now:         func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	return manifest
}

func saveBaseAndRemote(t *testing.T, root string, manifest Manifest) {
	t.Helper()
	hash, err := Hash(manifest)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if err := Save(root, State{
		LastRemoteManifestHash: hash,
		LastSyncAt:             time.Date(2026, 5, 3, 13, 0, 0, 0, time.UTC),
		Entries:                manifest.Entries,
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	writeRemoteManifest(t, root, manifest)
}

func writeRemoteManifest(t *testing.T, root string, manifest Manifest) {
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

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return content
}
