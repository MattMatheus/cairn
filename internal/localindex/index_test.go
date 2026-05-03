package localindex

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesLocalMetadataDatabase(t *testing.T) {
	root := t.TempDir()
	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()

	if _, err := os.Stat(DBPath(root)); err != nil {
		t.Fatalf("expected database at %s: %v", DBPath(root), err)
	}
}

func TestIndexWorkspaceIndexesManagedMarkdownAndSkipsInvalidFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdown("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth", "timeouts"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z"))
	writeFile(t, root, "notes/plain.md", "# Plain markdown\n")
	writeFile(t, root, "notes/invalid.md", "---\nid: bad\n---\n")
	writeFile(t, root, "notes/ignore.txt", "not markdown")

	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()

	report, err := index.IndexWorkspace(context.Background(), root)
	if err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}
	if len(report.Indexed) != 1 || report.Indexed[0] != "runbooks/auth.md" {
		t.Fatalf("unexpected indexed paths: %#v", report.Indexed)
	}
	if len(report.Skipped) != 2 {
		t.Fatalf("expected two skipped markdown files, got %#v", report.Skipped)
	}

	results, err := index.Query(context.Background(), Query{Slug: "auth-timeout-runbook"})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %#v", results)
	}
	result := results[0]
	if result.Path != "runbooks/auth.md" ||
		result.Title != "Auth Timeout Runbook" ||
		result.Type != "runbook" ||
		result.Status != "canonical" ||
		result.Slug != "auth-timeout-runbook" ||
		result.MatchType != MatchTypeMetadata ||
		result.Score != 1 ||
		result.Snippet == "" {
		t.Fatalf("result shape did not match ADR fields: %#v", result)
	}
	if len(result.Tags) != 2 || result.Tags[0] != "auth" || result.Tags[1] != "timeouts" {
		t.Fatalf("unexpected tags: %#v", result.Tags)
	}
	if result.Provenance.Source != "promotion" ||
		len(result.Provenance.Authors) != 1 || result.Provenance.Authors[0] != "matt" ||
		len(result.Provenance.Actors) != 1 || result.Provenance.Actors[0] != "codex" {
		t.Fatalf("unexpected provenance: %#v", result.Provenance)
	}
}

func TestQuerySupportsRepresentativeMetadataFiltersAndRecentOrdering(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdown("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z"))
	writeFile(t, root, "services/billing.md", managedMarkdown("cairn:billing", "Billing Service", "billing-service", "service", "proposed", []string{"billing"}, []string{"assistant"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z"))

	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}

	assertFirstPath(t, index, Query{Text: "Auth"}, "runbooks/auth.md")
	assertFirstPath(t, index, Query{Tag: "billing"}, "services/billing.md")
	assertFirstPath(t, index, Query{Status: "canonical"}, "runbooks/auth.md")
	assertFirstPath(t, index, Query{Type: "service"}, "services/billing.md")
	assertFirstPath(t, index, Query{Path: "services/billing.md"}, "services/billing.md")
	assertFirstPath(t, index, Query{Actor: "assistant"}, "services/billing.md")
	assertFirstPath(t, index, Query{Source: "promotion"}, "runbooks/auth.md")
	assertFirstPath(t, index, Query{Recent: true}, "services/billing.md")
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

func managedMarkdown(id string, title string, slug string, docType string, status string, tags []string, actors []string, authors []string, source string, updated string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + slug + `
type: ` + docType + `
status: ` + status + `
created: 2026-05-01T12:00:00Z
updated: ` + updated + `
authors: ` + inlineArray(authors) + `
actors: ` + inlineArray(actors) + `
source: ` + source + `
tags: ` + inlineArray(tags) + `
---

# ` + title + `
`
}

func inlineArray(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	result := "["
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result + "]"
}

func assertFirstPath(t *testing.T, index *Index, query Query, want string) {
	t.Helper()
	results, err := index.Query(context.Background(), query)
	if err != nil {
		t.Fatalf("Query(%#v) returned error: %v", query, err)
	}
	if len(results) == 0 {
		t.Fatalf("Query(%#v) returned no results", query)
	}
	if results[0].Path != want {
		t.Fatalf("Query(%#v) first path = %s, want %s; all results %#v", query, results[0].Path, want, results)
	}
}
