package mcpops

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
)

type Local struct {
	Root  string
	Index *localindex.Index
	Now   func() time.Time
}

func OpenLocal(root string) (*Local, error) {
	index, err := localindex.Open(root)
	if err != nil {
		return nil, err
	}
	return &Local{Root: root, Index: index}, nil
}

func (l *Local) Close() error {
	if l == nil || l.Index == nil {
		return nil
	}
	return l.Index.Close()
}

func (l *Local) GetBootstrap(_ context.Context, _ mcpschema.EmptyRequest) (mcpschema.Envelope[mcpschema.GetBootstrapData], error) {
	return mcpschema.Envelope[mcpschema.GetBootstrapData]{
		OK: true,
		Data: mcpschema.GetBootstrapData{
			WorkspaceID: filepath.Base(l.Root),
			Summary:     "Local Cairn workspace context is available.",
			NextSteps: []string{
				string(mcpschema.ToolSearchContext),
				string(mcpschema.ToolListDocuments),
				string(mcpschema.ToolIndexStatus),
			},
		},
		NextSteps: []mcpschema.NextStep{
			{Action: string(mcpschema.ToolSearchContext), Label: "Search local context"},
			{Action: string(mcpschema.ToolListDocuments), Label: "List managed documents"},
		},
		Provenance: l.provenance("bootstrap"),
	}, nil
}

func (l *Local) SearchContext(ctx context.Context, req mcpschema.SearchContextRequest) (mcpschema.Envelope[mcpschema.SearchContextData], error) {
	return l.Index.Search(ctx, l.Root, localindex.SearchOptions{
		Query: req.Query,
		Mode:  req.Mode,
		Limit: req.Limit,
	})
}

func (l *Local) ListDocuments(ctx context.Context, req mcpschema.ListDocumentsRequest) (mcpschema.Envelope[mcpschema.DocumentListData], error) {
	results, err := l.Index.Query(ctx, localindex.Query{
		Tag:    first(req.Tags),
		Status: req.Status,
		Type:   req.Type,
		Actor:  req.ActorFilter,
		Source: req.Source,
		Limit:  req.Limit,
		Recent: true,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.DocumentListData]{}, err
	}
	return mcpschema.Envelope[mcpschema.DocumentListData]{
		OK:         true,
		Data:       mcpschema.DocumentListData{Documents: summaries(results)},
		Provenance: l.provenance("local_index"),
	}, nil
}

func (l *Local) FindDocument(ctx context.Context, req mcpschema.FindDocumentRequest) (mcpschema.Envelope[mcpschema.DocumentListData], error) {
	results, err := l.Index.Query(ctx, localindex.Query{
		ID:     req.ID,
		Text:   req.Title,
		Slug:   req.Slug,
		Path:   req.Path,
		Type:   req.Type,
		Status: req.Status,
		Tag:    req.Tag,
		Limit:  req.Limit,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.DocumentListData]{}, err
	}
	return mcpschema.Envelope[mcpschema.DocumentListData]{
		OK:         true,
		Data:       mcpschema.DocumentListData{Documents: summaries(results)},
		Provenance: l.provenance("local_index"),
	}, nil
}

func (l *Local) IndexStatus(_ context.Context, _ mcpschema.IndexStatusRequest) (mcpschema.Envelope[mcpschema.IndexStatusData], error) {
	info, err := os.Stat(localindex.DBPath(l.Root))
	localAvailable := err == nil
	var lastRefresh time.Time
	if localAvailable {
		lastRefresh = info.ModTime().UTC().Truncate(time.Second)
	}
	envelope := mcpschema.Envelope[mcpschema.IndexStatusData]{
		OK: true,
		Data: mcpschema.IndexStatusData{
			LocalAvailable:  localAvailable,
			RemoteAvailable: false,
			Fresh:           localAvailable,
			LastRefreshAt:   lastRefresh,
		},
		Provenance: l.provenance("local_index"),
	}
	envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
		Code:    mcpschema.WarningRemoteService,
		Message: "remote indexer is unavailable in local profile",
	})
	envelope.Unavailable = append(envelope.Unavailable, mcpschema.UnavailableMode{
		Mode:      "remote",
		Reason:    mcpschema.WarningRemoteService,
		Message:   "remote indexer is not configured",
		Retryable: false,
	})
	envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
		Action: string(mcpschema.ToolSearchContext),
		Label:  "Use local search",
		Reason: "Local metadata and full-text search remain available.",
	})
	return envelope, nil
}

func (l *Local) provenance(source string) mcpschema.ResponseProvenance {
	now := time.Now().UTC()
	if l.Now != nil {
		now = l.Now().UTC()
	}
	return mcpschema.ResponseProvenance{
		Profile:     mcpschema.ProfileLocal,
		WorkspaceID: filepath.Base(l.Root),
		GeneratedAt: now,
		Source:      source,
	}
}

func summaries(results []mcpschema.SearchResult) []mcpschema.DocumentSummary {
	out := make([]mcpschema.DocumentSummary, 0, len(results))
	for _, result := range results {
		out = append(out, mcpschema.DocumentSummary{
			Path:    result.Path,
			Title:   result.Title,
			Slug:    result.Slug,
			Type:    result.Type,
			Status:  result.Status,
			Tags:    result.Tags,
			Updated: result.Updated,
			Authors: result.Provenance.Authors,
			Actors:  result.Provenance.Actors,
			Source:  result.Provenance.Source,
		})
	}
	return out
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
