package remoteindex

import (
	"context"
	"testing"
	"time"

	"cairn/internal/mcpschema"
)

func TestFakeClientMapsSearchToCairnEnvelope(t *testing.T) {
	updated := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	client := &FakeClient{
		SearchResponse: SearchResponse{
			AttemptedModes: []mcpschema.SearchMode{mcpschema.SearchModeSemantic},
			Results: []SearchResult{{
				Path:    "runbooks/auth-timeouts.md",
				Title:   "Auth Timeouts",
				Type:    "runbook",
				Status:  "canonical",
				Slug:    "auth-timeouts",
				Tags:    []string{"auth", "timeouts"},
				Updated: updated,
				Score:   0.91,
				Snippet: "Retry auth token requests with bounded exponential backoff.",
				Authors: []string{"matt"},
				Actors:  []string{"codex"},
				Source:  "promotion",
			}},
		},
	}
	response, err := client.Search(context.Background(), SearchRequest{WorkspaceID: "pod-1", Query: "auth timeout"})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	envelope := response.Envelope()
	if !envelope.OK {
		t.Fatalf("expected ok envelope")
	}
	if envelope.Provenance.Profile != mcpschema.ProfilePodRemote || envelope.Provenance.Source != "remote_indexer" {
		t.Fatalf("unexpected provenance %#v", envelope.Provenance)
	}
	if len(envelope.Data.Results) != 1 {
		t.Fatalf("unexpected results %#v", envelope.Data.Results)
	}
	result := envelope.Data.Results[0]
	if result.Path != "runbooks/auth-timeouts.md" || result.MatchType != mcpschema.SearchModeSemantic || result.Score != 0.91 {
		t.Fatalf("unexpected result %#v", result)
	}
	if result.Provenance.Source != "promotion" || result.Provenance.Authors[0] != "matt" {
		t.Fatalf("unexpected result provenance %#v", result.Provenance)
	}
	if client.Calls[0] != "search:pod-1:auth timeout" {
		t.Fatalf("unexpected calls %#v", client.Calls)
	}
}

func TestStatusMapsToIndexStatus(t *testing.T) {
	lastRefresh := time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC)
	status := StatusResponse{Available: true, Fresh: true, LastRefreshAt: lastRefresh, IndexedCount: 7}
	mapped := status.SchemaStatus()
	if mapped.LocalAvailable {
		t.Fatalf("remote status should not claim local availability")
	}
	if !mapped.RemoteAvailable || !mapped.Fresh || !mapped.LastRefreshAt.Equal(lastRefresh) {
		t.Fatalf("unexpected mapped status %#v", mapped)
	}
}
