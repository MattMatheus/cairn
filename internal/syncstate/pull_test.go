package syncstate

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplyPullCreatesAndEditsRemoteOnlyFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	store := newObjectStore()
	remoteContent := managedDocument("cairn:a", "working", "note", "A Remote")
	remote := manifestWithEntries(base, entryForContent("working/a.md", remoteContent, "cairn:a", "working", "note"))
	writeRemoteManifest(t, root, remote)
	writeRemoteObject(t, store, "working/a.md", remoteContent)

	status := mustStatus(t, root)
	_, err := ApplyPull(context.Background(), root, status, store, PullOptions{Now: fixedPullTime})
	if err != nil {
		t.Fatalf("ApplyPull() error = %v", err)
	}
	if got := string(readBytes(t, filepath.Join(root, "working", "a.md"))); !strings.Contains(got, "A Remote") {
		t.Fatalf("pull did not edit local file:\n%s", got)
	}
	assertStateMatchesRemote(t, root, remote)
}

func TestApplyPullMovesArchivesAndDeletes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	writeFile(t, root, "working/delete.md", managedDocument("cairn:delete", "working", "note", "Delete"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	store := newObjectStore()
	archivedContent := managedDocument("cairn:a", "archived", "note", "A")
	remote := Manifest{
		ManifestVersion: ManifestVersion,
		GeneratedAt:     base.GeneratedAt,
		WorkspaceID:     base.WorkspaceID,
		Entries: []Entry{
			entryForContent("archive/a.md", archivedContent, "cairn:a", "archived", "note"),
		},
	}
	writeRemoteManifest(t, root, remote)
	writeRemoteObject(t, store, "archive/a.md", archivedContent)

	status := mustStatus(t, root)
	_, err := ApplyPull(context.Background(), root, status, store, PullOptions{Now: fixedPullTime})
	if err != nil {
		t.Fatalf("ApplyPull() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "working", "a.md")); !os.IsNotExist(err) {
		t.Fatalf("old moved path still exists or stat failed differently: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "working", "delete.md")); !os.IsNotExist(err) {
		t.Fatalf("deleted path still exists or stat failed differently: %v", err)
	}
	if got := string(readBytes(t, filepath.Join(root, "archive", "a.md"))); !strings.Contains(got, "status: archived") {
		t.Fatalf("archive path was not written:\n%s", got)
	}
	assertStateMatchesRemote(t, root, remote)
}

func TestApplyPullRefusesDivergenceWithoutMutating(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A Local"))
	beforeState := readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))
	beforeFile := readBytes(t, filepath.Join(root, "working", "a.md"))

	remoteContent := managedDocument("cairn:a", "working", "note", "A Remote")
	remote := manifestWithEntries(base, entryForContent("working/a.md", remoteContent, "cairn:a", "working", "note"))
	writeRemoteManifest(t, root, remote)
	store := newObjectStore()
	writeRemoteObject(t, store, "working/a.md", remoteContent)

	status := mustStatus(t, root)
	if _, err := ApplyPull(context.Background(), root, status, store, PullOptions{Now: fixedPullTime}); err == nil {
		t.Fatalf("expected divergent pull to refuse")
	}
	if string(beforeState) != string(readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))) {
		t.Fatalf("refused pull mutated sync state")
	}
	if string(beforeFile) != string(readBytes(t, filepath.Join(root, "working", "a.md"))) {
		t.Fatalf("refused pull mutated local file")
	}
}

func TestApplyPullDoesNotAdvanceStateWhenObjectMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	beforeState := readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))

	remoteContent := managedDocument("cairn:remote", "working", "note", "Remote")
	remote := base
	remote.Entries = append(remote.Entries, entryForContent("working/remote.md", remoteContent, "cairn:remote", "working", "note"))
	writeRemoteManifest(t, root, remote)

	status := mustStatus(t, root)
	if _, err := ApplyPull(context.Background(), root, status, newObjectStore(), PullOptions{Now: fixedPullTime}); err == nil {
		t.Fatalf("expected missing remote object error")
	}
	if string(beforeState) != string(readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))) {
		t.Fatalf("failed pull advanced sync state")
	}
}

func TestApplyPullDoesNotRemoveMovedPathWhenObjectMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	remoteContent := managedDocument("cairn:a", "archived", "note", "A")
	remote := Manifest{
		ManifestVersion: ManifestVersion,
		GeneratedAt:     base.GeneratedAt,
		WorkspaceID:     base.WorkspaceID,
		Entries: []Entry{
			entryForContent("archive/a.md", remoteContent, "cairn:a", "archived", "note"),
		},
	}
	writeRemoteManifest(t, root, remote)

	status := mustStatus(t, root)
	if _, err := ApplyPull(context.Background(), root, status, newObjectStore(), PullOptions{Now: fixedPullTime}); err == nil {
		t.Fatalf("expected missing moved remote object error")
	}
	if _, err := os.Stat(filepath.Join(root, "working", "a.md")); err != nil {
		t.Fatalf("old moved path should remain after failed fetch: %v", err)
	}
}

func TestApplyPullRefusesInvalidRemoteMarkdownBeforeLocalWrites(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	beforeState := readBytes(t, statePath(root))

	remoteContent := "# Missing frontmatter\n"
	remote := base
	remote.Entries = append(remote.Entries, entryForContent("working/remote.md", remoteContent, "", "", ""))
	writeRemoteManifest(t, root, remote)
	store := newObjectStore()
	writeRemoteObject(t, store, "working/remote.md", remoteContent)

	status := mustStatus(t, root)
	_, err := ApplyPull(context.Background(), root, status, store, PullOptions{Now: fixedPullTime})
	if err == nil {
		t.Fatalf("expected validation refusal")
	}
	if _, ok := err.(ValidationError); !ok {
		t.Fatalf("expected ValidationError, got %T %v", err, err)
	}
	if _, err := os.Stat(filepath.Join(root, "working", "remote.md")); !os.IsNotExist(err) {
		t.Fatalf("invalid remote object should not be written, stat err=%v", err)
	}
	if string(beforeState) != string(readBytes(t, statePath(root))) {
		t.Fatalf("validation refusal advanced sync state")
	}
}

func mustStatus(t *testing.T, root string) Status {
	t.Helper()
	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	return status
}

func manifestWithEntries(base Manifest, replacements ...Entry) Manifest {
	remote := base
	byPath := entriesByPath(remote.Entries)
	for _, entry := range replacements {
		byPath[entry.Path] = entry
	}
	remote.Entries = remote.Entries[:0]
	for _, entry := range byPath {
		remote.Entries = append(remote.Entries, entry)
	}
	normalizeEntries(remote.Entries)
	return remote
}

func entryForContent(path string, content string, id string, status string, docType string) Entry {
	return Entry{
		Path:       path,
		Kind:       "file",
		Size:       int64(len(content)),
		Hash:       hashBytes([]byte(content)),
		Modified:   fixedPullTime(),
		DocumentID: id,
		Status:     status,
		Type:       docType,
	}
}

type objectStore struct {
	objects map[string][]byte
}

func newObjectStore() *objectStore {
	return &objectStore{objects: map[string][]byte{}}
}

func (s *objectStore) ReadObject(ctx context.Context, path string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	content, ok := s.objects[path]
	return append([]byte(nil), content...), ok, nil
}

func writeRemoteObject(t *testing.T, store *objectStore, path string, content string) {
	t.Helper()
	store.objects[path] = []byte(content)
}

func assertStateMatchesRemote(t *testing.T, root string, remote Manifest) {
	t.Helper()
	state, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	wantHash, err := Hash(remote)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if state.LastRemoteManifestHash != wantHash {
		t.Fatalf("unexpected state hash: got %s want %s", state.LastRemoteManifestHash, wantHash)
	}
	assertPaths(t, state.Entries, paths(remote.Entries))
}

func fixedPullTime() time.Time {
	return time.Date(2026, 5, 3, 14, 0, 0, 0, time.UTC)
}
