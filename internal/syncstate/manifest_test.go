package syncstate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateHonorsCairnignoreAndIncludesMetadata(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairnignore", "*.tmp\nbuild/\n/secrets.md\n")
	writeFile(t, root, "docs/keep.md", managedDocument("cairn:keep", "working", "note", "Keep"))
	writeFile(t, root, "notes/plain.md", "# Plain\n")
	writeFile(t, root, "scratch.tmp", "ignore me")
	writeFile(t, root, "build/output.txt", "ignore me")
	writeFile(t, root, "secrets.md", managedDocument("cairn:secret", "working", "note", "Secret"))

	manifest, err := Generate(root, GenerateOptions{
		WorkspaceID: "pod-1",
		Now:         func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	if manifest.ManifestVersion != ManifestVersion {
		t.Fatalf("unexpected manifest version: %d", manifest.ManifestVersion)
	}
	if manifest.WorkspaceID != "pod-1" {
		t.Fatalf("unexpected workspace id: %q", manifest.WorkspaceID)
	}
	assertPaths(t, manifest.Entries, []string{".cairnignore", "docs/keep.md", "notes/plain.md"})

	keep := entryByPath(t, manifest, "docs/keep.md")
	if keep.DocumentID != "cairn:keep" || keep.Status != "working" || keep.Type != "note" {
		t.Fatalf("expected document metadata on keep.md, got %#v", keep)
	}
	plain := entryByPath(t, manifest, "notes/plain.md")
	if plain.DocumentID != "" || plain.Status != "" || plain.Type != "" {
		t.Fatalf("expected plain markdown to omit document metadata, got %#v", plain)
	}
	if keep.Hash == "" || keep.Size == 0 || keep.Modified.IsZero() {
		t.Fatalf("expected content fields on keep.md, got %#v", keep)
	}
}

func TestStateRoundTripStoresNormalizedEntriesAndManifestHash(t *testing.T) {
	root := t.TempDir()
	lastSync := time.Date(2026, 5, 3, 13, 14, 15, 999, time.UTC)
	manifest := Manifest{
		ManifestVersion: ManifestVersion,
		WorkspaceID:     "pod-1",
		Entries: []Entry{
			{Path: "z.md", Kind: "file", Size: 1, Hash: "z", Modified: lastSync},
			{Path: "a.md", Kind: "file", Size: 1, Hash: "a", Modified: lastSync},
		},
	}
	hash, err := Hash(manifest)
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	err = Save(root, State{
		LastRemoteManifestHash: hash,
		LastSyncAt:             lastSync,
		Entries:                manifest.Entries,
	})
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load(root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.StateVersion != StateVersion {
		t.Fatalf("unexpected state version: %d", loaded.StateVersion)
	}
	if loaded.LastRemoteManifestHash != hash {
		t.Fatalf("unexpected manifest hash: %q", loaded.LastRemoteManifestHash)
	}
	if !loaded.LastSyncAt.Equal(lastSync.Truncate(time.Second)) {
		t.Fatalf("unexpected last sync time: %s", loaded.LastSyncAt)
	}
	assertPaths(t, loaded.Entries, []string{"a.md", "z.md"})
}

func TestChangesClassifiesCreateEditMoveArchiveAndDelete(t *testing.T) {
	base := Manifest{Entries: []Entry{
		{Path: "docs/edit.md", Kind: "file", Size: 1, Hash: "old", DocumentID: "cairn:edit", Status: "working", Type: "note"},
		{Path: "docs/move.md", Kind: "file", Size: 1, Hash: "same", DocumentID: "cairn:move", Status: "working", Type: "note"},
		{Path: "docs/archive.md", Kind: "file", Size: 1, Hash: "same", DocumentID: "cairn:archive", Status: "working", Type: "note"},
		{Path: "docs/delete.md", Kind: "file", Size: 1, Hash: "same", DocumentID: "cairn:delete", Status: "working", Type: "note"},
	}}
	current := Manifest{Entries: []Entry{
		{Path: "docs/edit.md", Kind: "file", Size: 2, Hash: "new", DocumentID: "cairn:edit", Status: "working", Type: "note"},
		{Path: "docs/moved.md", Kind: "file", Size: 1, Hash: "same", DocumentID: "cairn:move", Status: "working", Type: "note"},
		{Path: "archive/docs/archive.md", Kind: "file", Size: 2, Hash: "archived", DocumentID: "cairn:archive", Status: "archived", Type: "note"},
		{Path: "docs/create.md", Kind: "file", Size: 1, Hash: "new", DocumentID: "cairn:create", Status: "working", Type: "note"},
	}}

	changes := Changes(base, current)
	assertChange(t, changes, ChangeCreate, "docs/create.md")
	assertChange(t, changes, ChangeEdit, "docs/edit.md")
	assertChange(t, changes, ChangeMove, "docs/moved.md")
	assertChange(t, changes, ChangeArchive, "archive/docs/archive.md")
	assertChange(t, changes, ChangeDelete, "docs/delete.md")
}

func TestCompareAllowsSingleSidedChangesAndRefusesDivergence(t *testing.T) {
	base := Manifest{Entries: []Entry{{Path: "docs/a.md", Kind: "file", Size: 1, Hash: "base", DocumentID: "cairn:a", Status: "working", Type: "note"}}}
	local := Manifest{Entries: []Entry{{Path: "docs/a.md", Kind: "file", Size: 2, Hash: "local", DocumentID: "cairn:a", Status: "working", Type: "note"}}}
	remoteUnchanged := Manifest{Entries: []Entry{{Path: "docs/a.md", Kind: "file", Size: 1, Hash: "base", DocumentID: "cairn:a", Status: "working", Type: "note"}}}
	remoteChanged := Manifest{Entries: []Entry{{Path: "docs/a.md", Kind: "file", Size: 3, Hash: "remote", DocumentID: "cairn:a", Status: "working", Type: "note"}}}

	safe := Compare(base, local, remoteUnchanged)
	if safe.Diverged {
		t.Fatalf("expected single-sided local change to be allowed: %#v", safe)
	}
	if len(safe.LocalChanges) != 1 || len(safe.RemoteChanges) != 0 {
		t.Fatalf("unexpected safe comparison changes: %#v", safe)
	}

	refused := Compare(base, local, remoteChanged)
	if !refused.Diverged {
		t.Fatalf("expected local and remote changes to diverge: %#v", refused)
	}
	if len(refused.Conflicts) == 0 {
		t.Fatalf("expected refused comparison to include conflicts: %#v", refused)
	}
}

func writeFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	absolutePath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func managedDocument(id string, status string, docType string, title string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + testSlug(title) + `
type: ` + docType + `
status: ` + status + `
created: 2026-05-03T12:00:00Z
updated: 2026-05-03T12:00:00Z
authors: [matt]
actors: [codex]
source: capture
tags: []
---

# ` + title + `
`
}

func testSlug(value string) string {
	return strings.ToLower(strings.ReplaceAll(value, " ", "-"))
}

func assertPaths(t *testing.T, entries []Entry, want []string) {
	t.Helper()
	if len(entries) != len(want) {
		t.Fatalf("unexpected entry count: got %#v want %#v", paths(entries), want)
	}
	for i, entry := range entries {
		if entry.Path != want[i] {
			t.Fatalf("unexpected paths: got %#v want %#v", paths(entries), want)
		}
	}
}

func paths(entries []Entry) []string {
	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Path)
	}
	return got
}

func entryByPath(t *testing.T, manifest Manifest, path string) Entry {
	t.Helper()
	for _, entry := range manifest.Entries {
		if entry.Path == path {
			return entry
		}
	}
	t.Fatalf("missing entry for %s in %#v", path, manifest.Entries)
	return Entry{}
}

func assertChange(t *testing.T, changes []Change, changeType ChangeType, path string) {
	t.Helper()
	for _, change := range changes {
		if change.Type == changeType && change.Path == path {
			return
		}
	}
	t.Fatalf("missing change %s for %s in %#v", changeType, path, changes)
}
