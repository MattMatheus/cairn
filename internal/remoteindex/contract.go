package remoteindex

import (
	"context"
	"time"

	"cairn/internal/mcpschema"
)

type Client interface {
	Status(ctx context.Context, req StatusRequest) (StatusResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (RefreshResponse, error)
	Search(ctx context.Context, req SearchRequest) (SearchResponse, error)
}

type StatusRequest struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
}

type StatusResponse struct {
	Available     bool      `json:"available"`
	Fresh         bool      `json:"fresh"`
	LastRefreshAt time.Time `json:"last_refresh_at,omitempty"`
	IndexedCount  int       `json:"indexed_count,omitempty"`
	Message       string    `json:"message,omitempty"`
}

type RefreshRequest struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
	DryRun      bool   `json:"dry_run,omitempty"`
}

type RefreshResponse struct {
	Accepted      bool      `json:"accepted"`
	Refreshed     bool      `json:"refreshed"`
	JobID         string    `json:"job_id,omitempty"`
	LastRefreshAt time.Time `json:"last_refresh_at,omitempty"`
	Message       string    `json:"message,omitempty"`
}

type SearchRequest struct {
	WorkspaceID string               `json:"workspace_id,omitempty"`
	Query       string               `json:"query"`
	Mode        mcpschema.SearchMode `json:"mode,omitempty"`
	Types       []string             `json:"types,omitempty"`
	Tags        []string             `json:"tags,omitempty"`
	Limit       int                  `json:"limit,omitempty"`
}

type SearchResponse struct {
	Results        []SearchResult           `json:"results"`
	AttemptedModes []mcpschema.SearchMode   `json:"attempted_modes,omitempty"`
	Warnings       []mcpschema.Warning      `json:"warnings,omitempty"`
	NextSteps      []mcpschema.NextStep     `json:"next_steps,omitempty"`
	Provenance     mcpschema.ItemProvenance `json:"provenance,omitempty"`
}

type SearchResult struct {
	Path       string            `json:"path"`
	Title      string            `json:"title,omitempty"`
	Type       string            `json:"type,omitempty"`
	Status     string            `json:"status,omitempty"`
	Slug       string            `json:"slug,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Updated    time.Time         `json:"updated,omitempty"`
	Score      float64           `json:"score,omitempty"`
	Snippet    string            `json:"snippet,omitempty"`
	Authors    []string          `json:"authors,omitempty"`
	Actors     []string          `json:"actors,omitempty"`
	Source     string            `json:"source,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func (r SearchResponse) Envelope() mcpschema.Envelope[mcpschema.SearchContextData] {
	attempted := r.AttemptedModes
	if len(attempted) == 0 {
		attempted = []mcpschema.SearchMode{mcpschema.SearchModeSemantic}
	}
	results := make([]mcpschema.SearchResult, 0, len(r.Results))
	for _, result := range r.Results {
		results = append(results, result.schemaResult())
	}
	attemptedStrings := make([]string, 0, len(attempted))
	for _, mode := range attempted {
		attemptedStrings = append(attemptedStrings, string(mode))
	}
	return mcpschema.Envelope[mcpschema.SearchContextData]{
		OK: true,
		Data: mcpschema.SearchContextData{
			Results:        results,
			AttemptedModes: attempted,
		},
		Warnings:  r.Warnings,
		NextSteps: r.NextSteps,
		Provenance: mcpschema.ResponseProvenance{
			Profile:        mcpschema.ProfilePodRemote,
			Source:         "remote_indexer",
			AttemptedModes: attemptedStrings,
		},
	}
}

func (r SearchResult) schemaResult() mcpschema.SearchResult {
	return mcpschema.SearchResult{
		Path:      r.Path,
		Title:     r.Title,
		Type:      r.Type,
		Status:    r.Status,
		Slug:      r.Slug,
		Tags:      r.Tags,
		Updated:   r.Updated,
		Score:     r.Score,
		MatchType: mcpschema.SearchModeSemantic,
		Snippet:   r.Snippet,
		Provenance: mcpschema.ItemProvenance{
			Authors: r.Authors,
			Actors:  r.Actors,
			Source:  r.Source,
		},
	}
}

func (r StatusResponse) SchemaStatus() mcpschema.IndexStatusData {
	return mcpschema.IndexStatusData{
		LocalAvailable:  false,
		RemoteAvailable: r.Available,
		Fresh:           r.Fresh,
		LastRefreshAt:   r.LastRefreshAt,
	}
}
