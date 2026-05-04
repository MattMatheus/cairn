package syncstate

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestApplyPushWritesCreatesAndEditsThenManifest(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A Local"))
	writeFile(t, root, "working/local.md", managedDocument("cairn:local", "working", "note", "Local"))

	store := newPushStore()
	status := mustStatus(t, root)
	_, err := ApplyPush(context.Background(), root, status, store, PushOptions{WorkspaceID: "pod-1", Now: fixedPullTime})
	if err != nil {
		t.Fatalf("ApplyPush() error = %v", err)
	}
	if got := string(store.objects["working/a.md"]); !strings.Contains(got, "A Local") {
		t.Fatalf("edit was not pushed:\n%s", got)
	}
	if got := string(store.objects["working/local.md"]); !strings.Contains(got, "Local") {
		t.Fatalf("create was not pushed:\n%s", got)
	}
	if len(store.operations) < 3 || store.operations[len(store.operations)-1] != "write-manifest" {
		t.Fatalf("remote manifest was not published last: %#v", store.operations)
	}
	state, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	assertPaths(t, state.Entries, []string{"working/a.md", "working/local.md"})
}

func TestApplyPushMovesArchivesAndDeletesRemoteObjects(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	writeFile(t, root, "working/delete.md", managedDocument("cairn:delete", "working", "note", "Delete"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "archive/a.md", managedDocument("cairn:a", "archived", "note", "A"))
	if err := removeWorkspacePath(root, "working/a.md"); err != nil {
		t.Fatalf("removeWorkspacePath() error = %v", err)
	}
	if err := removeWorkspacePath(root, "working/delete.md"); err != nil {
		t.Fatalf("removeWorkspacePath() error = %v", err)
	}

	store := newPushStore()
	store.objects["working/a.md"] = []byte("old")
	store.objects["working/delete.md"] = []byte("delete")
	status := mustStatus(t, root)
	_, err := ApplyPush(context.Background(), root, status, store, PushOptions{WorkspaceID: "pod-1", Now: fixedPullTime})
	if err != nil {
		t.Fatalf("ApplyPush() error = %v", err)
	}
	if _, ok := store.objects["archive/a.md"]; !ok {
		t.Fatalf("archive object was not pushed")
	}
	if _, ok := store.objects["working/a.md"]; ok {
		t.Fatalf("previous moved object was not deleted")
	}
	if _, ok := store.objects["working/delete.md"]; ok {
		t.Fatalf("deleted object was not removed")
	}
}

func TestApplyPushRefusesDivergenceWithoutRemoteOrStateWrites(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A Local"))
	remoteContent := managedDocument("cairn:a", "working", "note", "A Remote")
	remote := manifestWithEntries(base, entryForContent("working/a.md", remoteContent, "cairn:a", "working", "note"))
	writeRemoteManifest(t, root, remote)
	beforeState := readBytes(t, statePath(root))

	store := newPushStore()
	status := mustStatus(t, root)
	if _, err := ApplyPush(context.Background(), root, status, store, PushOptions{WorkspaceID: "pod-1", Now: fixedPullTime}); err == nil {
		t.Fatalf("expected divergent push to refuse")
	}
	if len(store.operations) != 0 {
		t.Fatalf("refused push wrote remote operations: %#v", store.operations)
	}
	if string(beforeState) != string(readBytes(t, statePath(root))) {
		t.Fatalf("refused push mutated sync state")
	}
}

func TestApplyPushDoesNotPublishManifestOrAdvanceStateWhenObjectWriteFails(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "working/local.md", managedDocument("cairn:local", "working", "note", "Local"))
	beforeState := readBytes(t, statePath(root))

	store := newPushStore()
	store.failWriteObject = "working/local.md"
	status := mustStatus(t, root)
	if _, err := ApplyPush(context.Background(), root, status, store, PushOptions{WorkspaceID: "pod-1", Now: fixedPullTime}); err == nil {
		t.Fatalf("expected object write failure")
	}
	for _, op := range store.operations {
		if op == "write-manifest" {
			t.Fatalf("manifest was published after object write failure: %#v", store.operations)
		}
	}
	if string(beforeState) != string(readBytes(t, statePath(root))) {
		t.Fatalf("failed push advanced sync state")
	}
}

type pushStore struct {
	objects         map[string][]byte
	operations      []string
	failWriteObject string
}

func newPushStore() *pushStore {
	return &pushStore{objects: map[string][]byte{}}
}

func (s *pushStore) WriteManifest(ctx context.Context, manifest Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	s.operations = append(s.operations, "write-manifest")
	s.objects[".cairn/remote-manifest.json"] = append(content, '\n')
	return nil
}

func (s *pushStore) WriteObject(ctx context.Context, path string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.operations = append(s.operations, "write-object:"+path)
	if path == s.failWriteObject {
		return errors.New("write object failed")
	}
	s.objects[path] = append([]byte(nil), content...)
	return nil
}

func (s *pushStore) DeleteObject(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.operations = append(s.operations, "delete-object:"+path)
	delete(s.objects, path)
	return nil
}
