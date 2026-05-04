package remoteindex

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceImplementsRemoteIndexContract(t *testing.T) {
	root := t.TempDir()
	writeServiceFile(t, root, "runbooks/auth.md", managedServiceMarkdown("cairn:auth", "Auth Runbook", "auth-runbook", "runbook", "canonical", "Retry auth token requests."))
	service := NewService(root)
	service.Now = func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) }
	handler := service.Handler()

	var status StatusResponse
	postJSON(t, handler, "/index/status", StatusRequest{WorkspaceID: "pod-1"}, &status)
	if !status.Available || status.Fresh {
		t.Fatalf("unexpected initial status: %#v", status)
	}

	var refresh RefreshResponse
	postJSON(t, handler, "/index/refresh", RefreshRequest{WorkspaceID: "pod-1"}, &refresh)
	if !refresh.Accepted || !refresh.Refreshed || refresh.LastRefreshAt.IsZero() {
		t.Fatalf("unexpected refresh response: %#v", refresh)
	}

	var search SearchResponse
	postJSON(t, handler, "/search", SearchRequest{WorkspaceID: "pod-1", Query: "auth token", Limit: 10}, &search)
	if len(search.Results) != 1 {
		t.Fatalf("expected one result, got %#v", search)
	}
	result := search.Results[0]
	if result.Path != "runbooks/auth.md" || result.Title != "Auth Runbook" || result.Source != "capture" {
		t.Fatalf("service result lost stable shape: %#v", result)
	}

	postJSON(t, handler, "/index/status", StatusRequest{WorkspaceID: "pod-1"}, &status)
	if !status.Fresh || status.IndexedCount != 1 {
		t.Fatalf("unexpected refreshed status: %#v", status)
	}
}

func TestServiceRefreshDryRunDoesNotMarkFresh(t *testing.T) {
	root := t.TempDir()
	writeServiceFile(t, root, "notes/a.md", managedServiceMarkdown("cairn:a", "A", "a", "note", "working", "body"))
	service := NewService(root)
	handler := service.Handler()

	body, err := json.Marshal(RefreshRequest{WorkspaceID: "pod-1", DryRun: true})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/index/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected refresh status %d body=%s", rec.Code, rec.Body.String())
	}
	var status StatusResponse
	postJSON(t, handler, "/index/status", StatusRequest{}, &status)
	if status.Fresh || status.IndexedCount != 0 {
		t.Fatalf("dry-run refresh should not mark service fresh: %#v", status)
	}
}

func postJSON(t *testing.T, handler http.Handler, path string, in any, out any) {
	t.Helper()
	body, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s returned %d body=%s", path, rec.Code, rec.Body.String())
	}
	if err := json.NewDecoder(rec.Body).Decode(out); err != nil {
		t.Fatalf("Decode(%s) error = %v", path, err)
	}
}

func writeServiceFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func managedServiceMarkdown(id string, title string, slug string, docType string, status string, body string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + slug + `
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

` + body + `
`
}
