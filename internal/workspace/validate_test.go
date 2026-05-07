package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
	"cairn/internal/syncstate"
)

func TestValidateReportsDocumentFindings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/missing.md", "# Missing frontmatter\n")
	writeFile(t, root, "working/invalid.md", `---
id: bad
schema_version: 1
title: Invalid
slug: Invalid Slug
type: mystery
status: draft
created: 2026-05-03T00:00:00Z
updated: 2026-05-03T00:00:00Z
authors: [foundry]
actors: [codex]
source: test
tags: [test]
---
# Invalid
`)

	data, err := Validate(context.Background(), root, ValidateOptions{Mode: document.ValidationModeDiscovery})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !data.Healthy {
		t.Fatalf("discovery validation should remain healthy with warnings: %#v", data.Findings)
	}
	assertFinding(t, data.Findings, "working/missing.md", "", mcpschema.WarningValidation, "warning")
	assertFinding(t, data.Findings, "working/invalid.md", "", mcpschema.WarningValidation, "warning")
}

func TestValidateDoesNotWarnBeforeFirstIndexOrSync(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/good.md", validDoc("cairn:GoodDoc", "Good Doc", "good-doc"))

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !data.Healthy || len(data.Findings) != 0 {
		t.Fatalf("fresh local workspace should validate without index/sync warnings: %#v", data.Findings)
	}
}

func TestValidateUsesDurableBoundarySeverity(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/missing.md", "# Missing frontmatter\n")

	data, err := Validate(context.Background(), root, ValidateOptions{Mode: document.ValidationModeDurableBoundary})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if data.Healthy {
		t.Fatalf("durable validation should be unhealthy with blocking document findings")
	}
	assertFinding(t, data.Findings, "working/missing.md", "", mcpschema.WarningValidation, "error")
}

func TestValidateSkipsIgnoredPaths(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairnignore", "ignored/\n")
	writeFile(t, root, "ignored/bad.md", "# Ignored\n")
	writeFile(t, root, "working/good.md", validDoc("cairn:GoodDoc", "Good Doc", "good-doc"))
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !data.Healthy {
		t.Fatalf("expected healthy workspace, got findings: %#v", data.Findings)
	}
	if len(data.Findings) != 0 {
		t.Fatalf("expected ignored invalid document to produce no findings, got %#v", data.Findings)
	}
}

func TestValidateHealthyWorkspace(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/good.md", validDoc("cairn:GoodDoc", "Good Doc", "good-doc"))
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !data.Healthy {
		t.Fatalf("expected healthy workspace, got findings: %#v", data.Findings)
	}
	if len(data.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", data.Findings)
	}
}

func TestValidateRejectsRequestedPathsOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Dir(root)
	writeFile(t, parent, "outside.md", "# Outside\n")
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{Paths: []string{"../outside.md"}})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(data.Findings) != 0 {
		t.Fatalf("expected outside path to be skipped, got %#v", data.Findings)
	}
}

func TestValidateSkipsUnmanagedMarkdown(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "AGENTS.md", "# Pointer\n")
	writeFile(t, root, "docs/raw.md", "# Raw Doc\n")
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(data.Findings) != 0 {
		t.Fatalf("expected unmanaged markdown to be skipped, got %#v", data.Findings)
	}
}

func TestValidateUsesConfiguredManagedFolders(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
managed_folders:
  - knowledge/base
`)
	writeFile(t, root, "knowledge/base/raw.md", "# Missing frontmatter\n")
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	assertFinding(t, data.Findings, "knowledge/base/raw.md", "", mcpschema.WarningValidation, "warning")
}

func TestValidateSurfacesConfigAndSchemaFindings(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: nope
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  memo: working
`)
	writeFile(t, root, ".cairn/schemas/custom.yaml", `schema_version: 1
name: custom
required_fields:
  - id
  - title
`)
	writeHealthyMetadata(t, root)

	data, err := Validate(context.Background(), root, ValidateOptions{})
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if data.Healthy {
		t.Fatalf("config/schema errors should make validation unhealthy: %#v", data.Findings)
	}
	assertFinding(t, data.Findings, ".cairn/config.yaml", "", mcpschema.WarningValidation, "error")
	assertFinding(t, data.Findings, ".cairn/config.yaml", "", mcpschema.WarningValidation, "warning")
	assertFinding(t, data.Findings, ".cairn/schemas/custom.yaml", "", mcpschema.WarningValidation, "error")
}

func assertFinding(t *testing.T, findings []mcpschema.ValidationFinding, path string, docID string, code mcpschema.WarningCode, severity string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Path == path && finding.Code == code && finding.Severity == severity {
			if docID == "" || finding.DocumentID == docID {
				return
			}
		}
	}
	t.Fatalf("missing finding path=%q docID=%q code=%q severity=%q in %#v", path, docID, code, severity, findings)
}

func writeHealthyMetadata(t *testing.T, root string) {
	t.Helper()
	index, err := localindex.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := index.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := syncstate.Save(root, syncstate.State{
		LastRemoteManifestHash: "abc123",
		LastSyncAt:             time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
		Entries:                []syncstate.Entry{},
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func validDoc(id string, title string, slug string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + slug + `
type: note
status: draft
created: 2026-05-03T00:00:00Z
updated: 2026-05-03T00:00:00Z
authors: [foundry]
actors: [codex]
source: test
tags: [test]
---
# ` + title + `
`
}
