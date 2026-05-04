package localindex

import (
	"context"
	"errors"
	"testing"
	"time"

	"cairn/internal/mcpschema"
	"cairn/internal/remoteindex"
)

func TestFullTextSearchFindsManagedMarkdownBody(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdownWithBody("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z", "Retry auth token requests with bounded exponential backoff. Retry auth token requests once more."))
	writeFile(t, root, "runbooks/billing.md", managedMarkdownWithBody("cairn:billing", "Billing Runbook", "billing-runbook", "runbook", "canonical", []string{"billing"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T13:00:00Z", "Invoice reconciliation does not mention the phrase."))

	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}

	results, err := index.FullText(context.Background(), root, "auth token", 10)
	if err != nil {
		t.Fatalf("FullText returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one full-text result, got %#v", results)
	}
	result := results[0]
	if result.Path != "runbooks/auth.md" || result.MatchType != mcpschema.SearchModeFullText || result.Score != 2 {
		t.Fatalf("unexpected full-text result: %#v", result)
	}
	if result.Snippet == "" || result.Title != "Auth Timeout Runbook" || result.Provenance.Source != "promotion" {
		t.Fatalf("full-text result lost stable metadata fields: %#v", result)
	}
}

func TestFullTextSearchHonorsCairnignore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairnignore", "ignored/\n")
	writeFile(t, root, "ignored/hidden.md", managedMarkdownWithBody("cairn:hidden", "Hidden", "hidden", "note", "working", []string{"test"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "needle"))
	writeFile(t, root, "working/visible.md", managedMarkdownWithBody("cairn:visible", "Visible", "visible", "note", "working", []string{"test"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "needle"))

	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}

	results, err := index.FullText(context.Background(), root, "needle", 10)
	if err != nil {
		t.Fatalf("FullText returned error: %v", err)
	}
	if len(results) != 1 || results[0].Path != "working/visible.md" {
		t.Fatalf("unexpected full-text results: %#v", results)
	}
}

func TestSearchAutoUsesMetadataThenFullTextAndReportsDegradation(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdownWithBody("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z", "Retry auth token requests with bounded exponential backoff."))
	writeFile(t, root, "notes/token.md", managedMarkdownWithBody("cairn:token", "Credential Note", "credential-note", "note", "working", []string{"auth"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "This note discusses auth token rotation but not in its metadata."))

	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}

	envelope, err := index.Search(context.Background(), root, SearchOptions{Query: "auth", Mode: mcpschema.SearchModeAuto, Limit: 10})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if !envelope.OK {
		t.Fatalf("expected ok search envelope: %#v", envelope)
	}
	if len(envelope.Data.AttemptedModes) != 3 ||
		envelope.Data.AttemptedModes[0] != mcpschema.SearchModeMetadata ||
		envelope.Data.AttemptedModes[1] != mcpschema.SearchModeFullText ||
		envelope.Data.AttemptedModes[2] != mcpschema.SearchModeSemantic {
		t.Fatalf("unexpected attempted modes: %#v", envelope.Data.AttemptedModes)
	}
	if len(envelope.Data.Results) != 2 {
		t.Fatalf("expected metadata and full-text local results, got %#v", envelope.Data.Results)
	}
	if envelope.Data.Results[0].MatchType != mcpschema.SearchModeMetadata {
		t.Fatalf("expected metadata result first, got %#v", envelope.Data.Results)
	}
	if len(envelope.Warnings) != 2 || len(envelope.Unavailable) != 2 || len(envelope.NextSteps) == 0 {
		t.Fatalf("expected semantic/remote degradation reporting, got %#v", envelope)
	}
	if len(envelope.Provenance.AttemptedModes) != 3 {
		t.Fatalf("expected provenance attempted modes, got %#v", envelope.Provenance)
	}
}

func TestSemanticSearchModeDegradesGracefully(t *testing.T) {
	root := t.TempDir()
	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()

	envelope, err := index.Search(context.Background(), root, SearchOptions{Query: "anything", Mode: mcpschema.SearchModeSemantic})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if !envelope.OK || len(envelope.Data.Results) != 0 {
		t.Fatalf("expected graceful semantic degradation without results: %#v", envelope)
	}
	if len(envelope.Warnings) != 1 || len(envelope.Unavailable) != 1 || len(envelope.NextSteps) != 1 {
		t.Fatalf("expected degradation details, got %#v", envelope)
	}
}

func TestSemanticSearchModeUsesConfiguredRemoteIndexer(t *testing.T) {
	root := t.TempDir()
	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	remote := &remoteindex.FakeClient{
		SearchResponse: remoteindex.SearchResponse{
			Results: []remoteindex.SearchResult{{
				Path:    "runbooks/remote.md",
				Title:   "Remote Runbook",
				Type:    "runbook",
				Status:  "canonical",
				Slug:    "remote-runbook",
				Tags:    []string{"remote"},
				Updated: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
				Score:   0.92,
				Snippet: "semantic match",
				Authors: []string{"matt"},
				Actors:  []string{"codex"},
				Source:  "remote_index",
			}},
		},
	}

	envelope, err := index.Search(context.Background(), root, SearchOptions{
		Query:       "remote",
		Mode:        mcpschema.SearchModeSemantic,
		Limit:       10,
		WorkspaceID: "pod-1",
		Remote:      remote,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(remote.Calls) != 1 || remote.Calls[0] != "search:pod-1:remote" {
		t.Fatalf("remote search was not called correctly: %#v", remote.Calls)
	}
	if len(envelope.Data.Results) != 1 {
		t.Fatalf("expected remote result, got %#v", envelope)
	}
	result := envelope.Data.Results[0]
	if result.Path != "runbooks/remote.md" || result.MatchType != mcpschema.SearchModeSemantic || result.Provenance.Source != "remote_index" {
		t.Fatalf("remote result lost stable shape: %#v", result)
	}
	if len(envelope.Warnings) != 0 || len(envelope.Unavailable) != 0 {
		t.Fatalf("available remote search should not degrade: %#v", envelope)
	}
}

func TestSearchAutoMergesLocalAndRemoteResults(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdownWithBody("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z", "Retry auth token requests."))
	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}
	remote := &remoteindex.FakeClient{
		SearchResponse: remoteindex.SearchResponse{
			Results: []remoteindex.SearchResult{
				{Path: "runbooks/auth.md", Title: "Duplicate Remote", Score: 0.95, Source: "remote_index"},
				{Path: "runbooks/remote.md", Title: "Remote Only", Score: 0.9, Source: "remote_index"},
			},
		},
	}

	envelope, err := index.Search(context.Background(), root, SearchOptions{
		Query:       "auth",
		Mode:        mcpschema.SearchModeAuto,
		Limit:       10,
		WorkspaceID: "pod-1",
		Remote:      remote,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(envelope.Data.Results) != 2 {
		t.Fatalf("expected local result plus unique remote result, got %#v", envelope.Data.Results)
	}
	if envelope.Data.Results[0].Path != "runbooks/auth.md" || envelope.Data.Results[0].MatchType != mcpschema.SearchModeMetadata {
		t.Fatalf("expected local result to remain first: %#v", envelope.Data.Results)
	}
	if envelope.Data.Results[1].Path != "runbooks/remote.md" || envelope.Data.Results[1].MatchType != mcpschema.SearchModeSemantic {
		t.Fatalf("expected unique semantic remote result second: %#v", envelope.Data.Results)
	}
	if len(envelope.Warnings) != 0 || len(envelope.Unavailable) != 0 {
		t.Fatalf("configured remote search should not report degradation: %#v", envelope)
	}
}

func TestSearchAutoDegradesWhenRemoteIndexerFails(t *testing.T) {
	root := t.TempDir()
	index, err := Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	remote := failingRemote{err: errors.New("service unavailable")}

	envelope, err := index.Search(context.Background(), root, SearchOptions{
		Query:  "auth",
		Mode:   mcpschema.SearchModeAuto,
		Limit:  10,
		Remote: remote,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if !envelope.OK {
		t.Fatalf("search should gracefully degrade: %#v", envelope)
	}
	if len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != mcpschema.WarningRemoteService {
		t.Fatalf("expected remote warning, got %#v", envelope.Warnings)
	}
	if len(envelope.Unavailable) != 1 || len(envelope.NextSteps) != 1 {
		t.Fatalf("expected unavailable mode and next step, got %#v", envelope)
	}
}

type failingRemote struct {
	err error
}

func (f failingRemote) Status(ctx context.Context, req remoteindex.StatusRequest) (remoteindex.StatusResponse, error) {
	return remoteindex.StatusResponse{}, f.err
}

func (f failingRemote) Refresh(ctx context.Context, req remoteindex.RefreshRequest) (remoteindex.RefreshResponse, error) {
	return remoteindex.RefreshResponse{}, f.err
}

func (f failingRemote) Search(ctx context.Context, req remoteindex.SearchRequest) (remoteindex.SearchResponse, error) {
	return remoteindex.SearchResponse{}, f.err
}

func managedMarkdownWithBody(id string, title string, slug string, docType string, status string, tags []string, actors []string, authors []string, source string, updated string, body string) string {
	return managedMarkdown(id, title, slug, docType, status, tags, actors, authors, source, updated) + "\n" + body + "\n"
}
