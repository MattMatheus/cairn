package localindex

import (
	"context"
	"testing"

	"cairn/internal/mcpschema"
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

func managedMarkdownWithBody(id string, title string, slug string, docType string, status string, tags []string, actors []string, authors []string, source string, updated string, body string) string {
	return managedMarkdown(id, title, slug, docType, status, tags, actors, authors, source, updated) + "\n" + body + "\n"
}
