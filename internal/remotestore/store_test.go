package remotestore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/syncstate"
)

func TestMemoryStoreManifestAndObjects(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	manifest := syncstate.Manifest{
		ManifestVersion: syncstate.ManifestVersion,
		GeneratedAt:     time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
		WorkspaceID:     "pod-1",
		Entries: []syncstate.Entry{{
			Path:       "working/a.md",
			Kind:       "file",
			Size:       1,
			Hash:       "abc",
			DocumentID: "cairn:a",
			Status:     "working",
			Type:       "note",
		}},
	}

	if _, ok, err := store.ReadManifest(ctx); err != nil || ok {
		t.Fatalf("empty ReadManifest() = ok %t err %v", ok, err)
	}
	if err := store.WriteManifest(ctx, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	got, ok, err := store.ReadManifest(ctx)
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}
	if !ok || got.WorkspaceID != "pod-1" || len(got.Entries) != 1 {
		t.Fatalf("unexpected manifest ok=%t manifest=%#v", ok, got)
	}

	if err := store.WriteObject(ctx, "/working/a.md", []byte("hello")); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
	}
	content, ok, err := store.ReadObject(ctx, "working/a.md")
	if err != nil || !ok || string(content) != "hello" {
		t.Fatalf("ReadObject() = %q ok %t err %v", string(content), ok, err)
	}
	objects, err := store.ListObjects(ctx, "working")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}
	if len(objects) != 1 || objects[0].Path != "working/a.md" || objects[0].Size != 5 {
		t.Fatalf("unexpected objects %#v", objects)
	}
}

func TestLocalFSStoreManifestAndObjects(t *testing.T) {
	ctx := context.Background()
	store, err := NewLocalFSStore(t.TempDir(), "pod-a")
	if err != nil {
		t.Fatalf("NewLocalFSStore() error = %v", err)
	}
	manifest := syncstate.Manifest{
		ManifestVersion: syncstate.ManifestVersion,
		GeneratedAt:     time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
		WorkspaceID:     "pod-1",
		Entries: []syncstate.Entry{{
			Path: "working/a file.md",
			Kind: "file",
			Size: 5,
			Hash: "abc",
		}},
	}
	if err := store.WriteManifest(ctx, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	if err := store.WriteObject(ctx, "working/a file.md", []byte("hello")); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
	}
	content, ok, err := store.ReadObject(ctx, "working/a file.md")
	if err != nil || !ok || string(content) != "hello" {
		t.Fatalf("ReadObject() = %q ok=%t err=%v", string(content), ok, err)
	}
	objects, err := store.ListObjects(ctx, "working")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}
	if len(objects) != 1 || objects[0].Path != "working/a file.md" || objects[0].Size != 5 {
		t.Fatalf("unexpected objects %#v", objects)
	}
	got, ok, err := store.ReadManifest(ctx)
	if err != nil || !ok || got.WorkspaceID != "pod-1" {
		t.Fatalf("ReadManifest() ok=%t got=%#v err=%v", ok, got, err)
	}
	if _, err := os.Stat(filepath.Join(store.Root, "pod-a", ".cairn", "remote-manifest.json")); err != nil {
		t.Fatalf("manifest not written under prefix: %v", err)
	}
}

func TestLocalFSStoreRejectsNestedTraversal(t *testing.T) {
	ctx := context.Background()
	store, err := NewLocalFSStore(t.TempDir(), "")
	if err != nil {
		t.Fatalf("NewLocalFSStore() error = %v", err)
	}

	if err := store.WriteObject(ctx, "a/../../outside", []byte("nope")); err == nil || !strings.Contains(err.Error(), "unsafe local_fs object path") {
		t.Fatalf("expected unsafe path error, got %v", err)
	}
}

func TestPathMapping(t *testing.T) {
	if got := JoinPrefix("pod-a", "/working/a.md"); got != "pod-a/working/a.md" {
		t.Fatalf("JoinPrefix() = %q", got)
	}
	if got := JoinPrefix("", "working/a.md"); got != "working/a.md" {
		t.Fatalf("JoinPrefix() without prefix = %q", got)
	}
	if got := StripPrefix("pod-a", "pod-a/working/a.md"); got != "working/a.md" {
		t.Fatalf("StripPrefix() = %q", got)
	}
	if got := CleanPath("/a//b/../c.md"); got != "a/c.md" {
		t.Fatalf("CleanPath() = %q", got)
	}
}
