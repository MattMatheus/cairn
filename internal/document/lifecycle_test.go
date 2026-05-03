package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureCreatesManagedDocument(t *testing.T) {
	workspace := testWorkspace(t)

	result, err := workspace.Capture(CaptureOptions{
		Actor: "codex",
		Title: "Debug Auth Timeout",
		Body:  "# Debug Auth Timeout\n\nNotes.\n",
		Type:  "investigation",
		Tags:  []string{"auth"},
	})
	if err != nil {
		t.Fatalf("Capture returned error: %v", err)
	}
	if result.Path != filepath.Join("agents", "codex", "debug-auth-timeout.md") {
		t.Fatalf("unexpected path: %s", result.Path)
	}
	if result.DocumentID != "cairn:testid" {
		t.Fatalf("unexpected document id: %s", result.DocumentID)
	}

	parsed := mustParseFile(t, workspace, result.Path)
	if parsed.Metadata.Status != "working" {
		t.Fatalf("unexpected status: %s", parsed.Metadata.Status)
	}
	if parsed.Metadata.Type != "investigation" {
		t.Fatalf("unexpected type: %s", parsed.Metadata.Type)
	}
	if parsed.Metadata.Title != "Debug Auth Timeout" {
		t.Fatalf("unexpected title: %s", parsed.Metadata.Title)
	}
	if validation := Validate(parsed, ValidationModeDurableBoundary); validation.Blocking() {
		t.Fatalf("captured document should be valid: %#v", validation.Findings)
	}
}

func TestCaptureRejectsInvalidCoreFrontmatter(t *testing.T) {
	tests := []struct {
		name string
		opts CaptureOptions
	}{
		{
			name: "invalid type",
			opts: CaptureOptions{
				Actor: "codex",
				Title: "Bad Type",
				Type:  "not-a-type",
			},
		},
		{
			name: "empty slug",
			opts: CaptureOptions{
				Actor: "codex",
				Title: "!!!",
			},
		},
		{
			name: "invalid tag",
			opts: CaptureOptions{
				Actor: "codex",
				Title: "Bad Tag",
				Tags:  []string{"#inline"},
			},
		},
		{
			name: "unsafe actor parent path",
			opts: CaptureOptions{
				Actor: filepath.Join("..", "codex"),
				Title: "Unsafe Actor",
			},
		},
		{
			name: "unsafe actor windows separator",
			opts: CaptureOptions{
				Actor: `team\codex`,
				Title: "Unsafe Actor",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := testWorkspace(t)
			if _, err := workspace.Capture(test.opts); err == nil {
				t.Fatal("expected capture to reject invalid input")
			}
			assertNoMarkdownFiles(t, workspace.Root)
		})
	}
}

func TestPromoteToProposedRepairsMissingFrontmatter(t *testing.T) {
	workspace := testWorkspace(t)
	writeFile(t, workspace, filepath.Join("working", "rough-idea.md"), "# Rough Idea\n\nBody.\n")

	result, err := workspace.Promote(PromoteOptions{
		Path:   filepath.Join("working", "rough-idea.md"),
		Type:   "note",
		Status: "proposed",
	})
	if err != nil {
		t.Fatalf("Promote returned error: %v", err)
	}
	if result.Path != filepath.Join("working", "rough-idea.md") {
		t.Fatalf("unexpected target path: %s", result.Path)
	}

	parsed := mustParseFile(t, workspace, result.Path)
	if parsed.Metadata.ID != "cairn:testid" {
		t.Fatalf("expected repaired id, got %s", parsed.Metadata.ID)
	}
	if parsed.Metadata.Status != "proposed" {
		t.Fatalf("expected proposed status, got %s", parsed.Metadata.Status)
	}
	if parsed.Metadata.Title != "Rough Idea" {
		t.Fatalf("expected title from path, got %s", parsed.Metadata.Title)
	}
	if validation := Validate(parsed, ValidationModeDurableBoundary); validation.Blocking() {
		t.Fatalf("promoted document should be valid: %#v", validation.Findings)
	}
}

func assertNoMarkdownFiles(t *testing.T, root string) {
	t.Helper()
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			t.Fatalf("unexpected markdown file written: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatalf("WalkDir returned error: %v", err)
	}
}

func TestPromoteToProposedPreservesDurableMetadata(t *testing.T) {
	workspace := testWorkspace(t)
	sourcePath := filepath.Join("agents", "codex", "investigation.md")
	writeFile(t, workspace, sourcePath, validLifecycleDocument("cairn:preserveid", "Investigation", "investigation", "investigation", "working"))

	result, err := workspace.Promote(PromoteOptions{
		Path:   sourcePath,
		Status: "proposed",
	})
	if err != nil {
		t.Fatalf("Promote returned error: %v", err)
	}
	expectedPath := filepath.Join("working", "investigation.md")
	if result.Path != expectedPath {
		t.Fatalf("unexpected target path: %s", result.Path)
	}

	parsed := mustParseFile(t, workspace, expectedPath)
	if parsed.Metadata.ID != "cairn:preserveid" {
		t.Fatalf("document id was not preserved: %s", parsed.Metadata.ID)
	}
	if len(parsed.Metadata.Authors) != 1 || parsed.Metadata.Authors[0] != "matt" {
		t.Fatalf("authors were not preserved: %#v", parsed.Metadata.Authors)
	}
	if len(parsed.Metadata.Actors) != 1 || parsed.Metadata.Actors[0] != "codex" {
		t.Fatalf("actors were not preserved: %#v", parsed.Metadata.Actors)
	}
	if parsed.Metadata.Source != "capture" {
		t.Fatalf("source was not preserved: %s", parsed.Metadata.Source)
	}
}

func TestCanonicalPromotionBlocksInvalidFrontmatter(t *testing.T) {
	workspace := testWorkspace(t)
	writeFile(t, workspace, filepath.Join("working", "invalid.md"), `---
id: cairn:bad/id
schema_version: 1
title: Invalid
slug: invalid
type: decision
status: proposed
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors: [matt]
actors: [codex]
source: capture
tags: []
---

Body.
`)

	_, err := workspace.Promote(PromoteOptions{
		Path:   filepath.Join("working", "invalid.md"),
		Status: "canonical",
	})
	if err == nil {
		t.Fatal("expected canonical promotion to block invalid frontmatter")
	}
}

func TestCanonicalDecisionPromotionAssignsADRNumber(t *testing.T) {
	workspace := testWorkspace(t)
	writeFile(t, workspace, filepath.Join("decisions", "ADR-0007-existing.md"), "# Existing\n")
	writeFile(t, workspace, filepath.Join("working", "choose-storage.md"), validLifecycleDocument("cairn:decisionid", "Choose Storage", "choose-storage", "decision", "proposed"))

	result, err := workspace.Promote(PromoteOptions{
		Path:   filepath.Join("working", "choose-storage.md"),
		Status: "canonical",
	})
	if err != nil {
		t.Fatalf("Promote returned error: %v", err)
	}
	expectedPath := filepath.Join("decisions", "ADR-0008-choose-storage.md")
	if result.Path != expectedPath {
		t.Fatalf("unexpected path: %s", result.Path)
	}
	if _, err := os.Stat(filepath.Join(workspace.Root, filepath.Join("working", "choose-storage.md"))); !os.IsNotExist(err) {
		t.Fatalf("expected original file removed, stat err: %v", err)
	}

	parsed := mustParseFile(t, workspace, expectedPath)
	if parsed.Metadata.ID != "cairn:decisionid" {
		t.Fatalf("document id was not preserved: %s", parsed.Metadata.ID)
	}
	if parsed.Metadata.Status != "canonical" {
		t.Fatalf("unexpected status: %s", parsed.Metadata.Status)
	}
}

func TestArchivePreservesOriginalPathUnderArchive(t *testing.T) {
	workspace := testWorkspace(t)
	sourcePath := filepath.Join("decisions", "ADR-0001-old-choice.md")
	writeFile(t, workspace, sourcePath, validLifecycleDocument("cairn:archiveid", "Old Choice", "old-choice", "decision", "canonical"))

	result, err := workspace.Archive(ArchiveOptions{Path: sourcePath})
	if err != nil {
		t.Fatalf("Archive returned error: %v", err)
	}
	expectedPath := filepath.Join("archive", "decisions", "ADR-0001-old-choice.md")
	if result.Path != expectedPath {
		t.Fatalf("unexpected archive path: %s", result.Path)
	}
	if _, err := os.Stat(filepath.Join(workspace.Root, sourcePath)); !os.IsNotExist(err) {
		t.Fatalf("expected original file removed, stat err: %v", err)
	}

	parsed := mustParseFile(t, workspace, expectedPath)
	if parsed.Metadata.ID != "cairn:archiveid" {
		t.Fatalf("document id was not preserved: %s", parsed.Metadata.ID)
	}
	if parsed.Metadata.Status != "archived" {
		t.Fatalf("unexpected status: %s", parsed.Metadata.Status)
	}
}

func TestLifecycleOperationsRejectPathsOutsideWorkspace(t *testing.T) {
	workspace := testWorkspace(t)

	if _, err := workspace.Promote(PromoteOptions{Path: filepath.Join("..", "outside.md")}); err == nil {
		t.Fatal("expected promote to reject path traversal")
	}
	if _, err := workspace.Archive(ArchiveOptions{Path: filepath.Join("..", "outside.md")}); err == nil {
		t.Fatal("expected archive to reject path traversal")
	}
	if _, err := workspace.Promote(PromoteOptions{Path: filepath.Join(workspace.Root, "absolute.md")}); err == nil {
		t.Fatal("expected promote to reject absolute path")
	}
}

func testWorkspace(t *testing.T) Workspace {
	t.Helper()
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	return Workspace{
		Root: t.TempDir(),
		Now: func() time.Time {
			return now
		},
		NewID: func() string {
			return "cairn:testid"
		},
	}
}

func writeFile(t *testing.T, workspace Workspace, path string, content string) {
	t.Helper()
	absolutePath := filepath.Join(workspace.Root, path)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func mustParseFile(t *testing.T, workspace Workspace, path string) ParseResult {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(workspace.Root, path))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	parsed, err := ParseMarkdown(string(content))
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	return parsed
}

func validLifecycleDocument(id string, title string, slug string, docType string, status string) string {
	return strings.TrimLeft(`---
id: `+id+`
schema_version: 1
title: `+title+`
slug: `+slug+`
type: `+docType+`
status: `+status+`
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors: [matt]
actors: [codex]
source: capture
tags: []
---

Body.
`, "\n")
}
