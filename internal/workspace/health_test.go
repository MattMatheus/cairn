package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildAndRenderHealthReport(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root, InitOptions{WorkspaceID: "cairn:workspace:health"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	writeHealthDoc(t, root, "working/stale.md", "cairn:stale", "Stale Working", "note", "working", "2026-03-01T00:00:00Z")
	writeHealthDoc(t, root, "working/proposed.md", "cairn:proposed", "Proposed Doc", "runbook", "proposed", "2026-05-07T00:00:00Z")
	writeHealthDoc(t, root, "decisions/ADR-0001-choice.md", "cairn:choice", "Choice", "decision", "canonical", "2026-05-08T00:00:00Z")
	if err := os.WriteFile(filepath.Join(root, "runbooks", "bad.md"), []byte("# Missing Frontmatter\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	report, err := BuildHealthReport(context.Background(), root, HealthOptions{
		Now: func() time.Time { return time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("BuildHealthReport() error = %v", err)
	}
	if report.TotalManaged != 6 {
		t.Fatalf("unexpected total managed count: %d", report.TotalManaged)
	}
	if report.CountsByStatus["working"] != 1 || report.CountsByStatus["proposed"] != 1 || report.CountsByStatus["canonical"] != 1 {
		t.Fatalf("unexpected status counts: %#v", report.CountsByStatus)
	}
	if len(report.Proposed) != 1 || len(report.StaleWorking) != 1 || len(report.RecentCanonical) != 1 {
		t.Fatalf("unexpected health sections: proposed=%d stale=%d canonical=%d", len(report.Proposed), len(report.StaleWorking), len(report.RecentCanonical))
	}
	if len(report.ValidationFindings) == 0 {
		t.Fatalf("expected validation findings")
	}

	rendered := RenderHealthReport(report)
	for _, expected := range []string{
		"# Cairn Knowledge Health",
		"Managed documents: 6",
		"## Documents By Status",
		"- proposed: 1",
		"## Proposed Documents Awaiting Review",
		"Proposed Doc (`working/proposed.md`)",
		"## Stale Working Documents",
		"Stale Working (`working/stale.md`)",
		"## Validation Findings",
		"bad.md",
		"- Local index: missing",
		"- Next: run `cairn index refresh`.",
		"- Review proposed documents and promote or archive them.",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered report missing %q:\n%s", expected, rendered)
		}
	}
}

func writeHealthDoc(t *testing.T, root string, rel string, id string, title string, docType string, status string, updated string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + strings.TrimSuffix(filepath.Base(rel), ".md") + `
type: ` + docType + `
status: ` + status + `
created: 2026-03-01T00:00:00Z
updated: ` + updated + `
authors:
  - tester
actors:
  - tester
source: test
tags: []
---

# ` + title + `
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
