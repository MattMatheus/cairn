package mcpops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
)

func TestLocalSearchContextUsesEnvelopeAndDegradation(t *testing.T) {
	ops := newFixtureOps(t)
	envelope, err := ops.SearchContext(context.Background(), mcpschema.SearchContextRequest{
		Query: "auth",
		Mode:  mcpschema.SearchModeAuto,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchContext returned error: %v", err)
	}
	if !envelope.OK || len(envelope.Data.Results) == 0 {
		t.Fatalf("expected search results envelope: %#v", envelope)
	}
	if len(envelope.Warnings) != 2 || len(envelope.Unavailable) != 2 || len(envelope.NextSteps) == 0 {
		t.Fatalf("expected local-only degradation details: %#v", envelope)
	}
	if envelope.Provenance.Profile != mcpschema.ProfileLocal {
		t.Fatalf("expected local provenance: %#v", envelope.Provenance)
	}
}

func TestListAndFindDocumentsUseLocalMetadata(t *testing.T) {
	ops := newFixtureOps(t)
	listed, err := ops.ListDocuments(context.Background(), mcpschema.ListDocumentsRequest{
		Type:  "runbook",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListDocuments returned error: %v", err)
	}
	if len(listed.Data.Documents) != 1 || listed.Data.Documents[0].Path != "runbooks/auth.md" {
		t.Fatalf("unexpected list response: %#v", listed)
	}

	for name, req := range map[string]mcpschema.FindDocumentRequest{
		"id":     {ID: "cairn:auth"},
		"slug":   {Slug: "auth-timeout-runbook"},
		"title":  {Title: "Auth Timeout"},
		"path":   {Path: "runbooks/auth.md"},
		"type":   {Type: "runbook"},
		"status": {Status: "canonical"},
		"tag":    {Tag: "auth"},
	} {
		found, err := ops.FindDocument(context.Background(), req)
		if err != nil {
			t.Fatalf("FindDocument(%s) returned error: %v", name, err)
		}
		if len(found.Data.Documents) == 0 || found.Data.Documents[0].Path != "runbooks/auth.md" {
			t.Fatalf("FindDocument(%s) unexpected response: %#v", name, found)
		}
	}
}

func TestIndexStatusAndBootstrapAreLocalOnly(t *testing.T) {
	ops := newFixtureOps(t)
	status, err := ops.IndexStatus(context.Background(), mcpschema.IndexStatusRequest{})
	if err != nil {
		t.Fatalf("IndexStatus returned error: %v", err)
	}
	if !status.Data.LocalAvailable || status.Data.RemoteAvailable || !status.Data.Fresh {
		t.Fatalf("unexpected index status: %#v", status)
	}
	if len(status.Warnings) != 1 || len(status.Unavailable) != 1 || len(status.NextSteps) != 1 {
		t.Fatalf("expected local-only index degradation details: %#v", status)
	}

	bootstrap, err := ops.GetBootstrap(context.Background(), mcpschema.EmptyRequest{})
	if err != nil {
		t.Fatalf("GetBootstrap returned error: %v", err)
	}
	if !bootstrap.OK || bootstrap.Data.Summary == "" || len(bootstrap.NextSteps) == 0 {
		t.Fatalf("unexpected bootstrap: %#v", bootstrap)
	}
}

func newFixtureOps(t *testing.T) *Local {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdown("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z", "Retry auth token requests with bounded exponential backoff."))
	writeFile(t, root, "notes/token.md", managedMarkdown("cairn:token", "Credential Note", "credential-note", "note", "working", []string{"auth"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Token rotation note body."))
	index, err := localindex.Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() { index.Close() })
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}
	return &Local{
		Root:  root,
		Index: index,
		Now:   func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) },
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

func managedMarkdown(id string, title string, slug string, docType string, status string, tags []string, actors []string, authors []string, source string, updated string, body string) string {
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

` + body + `
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
