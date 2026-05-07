package mcpops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
	"cairn/internal/remoteindex"
	"cairn/internal/remotestore"
)

type Local struct {
	Root        string
	Index       *localindex.Index
	Now         func() time.Time
	RemoteIndex remoteindex.Client
	RemoteStore remotestore.Store
}

func OpenLocal(root string) (*Local, error) {
	index, err := localindex.Open(root)
	if err != nil {
		return nil, err
	}
	local := &Local{Root: root, Index: index}
	if err := configureRemoteClients(local); err != nil {
		index.Close()
		return nil, err
	}
	return local, nil
}

func (l *Local) Close() error {
	if l == nil || l.Index == nil {
		return nil
	}
	return l.Index.Close()
}

func configureRemoteClients(local *Local) error {
	cfg, err := document.LoadConfig(local.Root)
	if err != nil {
		return err
	}
	if remoteStore, err := remoteStoreFromConfig(local.Root, cfg.RemoteSync); err != nil {
		return err
	} else {
		local.RemoteStore = remoteStore
	}
	if cfg.RemoteIndex.URL != "" {
		local.RemoteIndex = remoteindex.HTTPClient{
			BaseURL: cfg.RemoteIndex.URL,
			Token:   remoteindex.AzureCLIToken(cfg.RemoteIndex.Audience),
		}
	}
	return nil
}

func remoteStoreFromConfig(root string, cfg document.RemoteSyncConfig) (remotestore.Store, error) {
	if cfg.Provider == "" && cfg.Account == "" && cfg.Endpoint == "" && cfg.Container == "" && cfg.Root == "" {
		return nil, nil
	}
	switch cfg.Provider {
	case "", "azure_blob":
		return remotestore.NewAzureBlobStore(remotestore.AzureBlobConfig{
			Account:   cfg.Account,
			Endpoint:  cfg.Endpoint,
			Container: cfg.Container,
			Prefix:    cfg.Prefix,
			AuthMode:  cfg.AuthMode,
		}, nil)
	case "local_fs":
		storeRoot := cfg.Root
		if storeRoot != "" && !filepath.IsAbs(storeRoot) {
			storeRoot = filepath.Join(root, filepath.FromSlash(storeRoot))
		}
		return remotestore.NewLocalFSStore(storeRoot, cfg.Prefix)
	default:
		return nil, fmt.Errorf("unsupported remote_sync provider %q", cfg.Provider)
	}
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
		Query:       req.Query,
		Mode:        req.Mode,
		Limit:       req.Limit,
		WorkspaceID: l.provenance("search_context").WorkspaceID,
		Remote:      l.RemoteIndex,
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

func (l *Local) IndexStatus(ctx context.Context, _ mcpschema.IndexStatusRequest) (mcpschema.Envelope[mcpschema.IndexStatusData], error) {
	info, err := os.Stat(localindex.DBPath(l.Root))
	localAvailable := err == nil
	var lastRefresh time.Time
	if localAvailable {
		lastRefresh = info.ModTime().UTC().Truncate(time.Second)
	}
	envelope := mcpschema.Envelope[mcpschema.IndexStatusData]{
		OK: true,
		Data: mcpschema.IndexStatusData{
			LocalAvailable: localAvailable,
			Fresh:          localAvailable,
			LastRefreshAt:  lastRefresh,
		},
		Provenance: l.provenance("local_index"),
	}
	if l.RemoteIndex == nil {
		if localAvailable {
			envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
				Action: string(mcpschema.ToolSearchContext),
				Label:  "Use local search",
				Reason: "Local metadata and full-text search are available.",
			})
		} else {
			envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
				Action: string(mcpschema.ToolIndexRefresh),
				Label:  "Refresh local index",
				Reason: "Create the local SQLite index before searching.",
			})
		}
		return envelope, nil
	}
	remote, err := l.RemoteIndex.Status(ctx, remoteindex.StatusRequest{
		WorkspaceID: l.provenance("index_status").WorkspaceID,
	})
	if err != nil {
		envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
			Code:    mcpschema.WarningRemoteService,
			Message: "remote indexer status failed: " + err.Error(),
		})
		envelope.Unavailable = append(envelope.Unavailable, mcpschema.UnavailableMode{
			Mode:      "remote",
			Reason:    mcpschema.WarningRemoteService,
			Message:   "remote indexer unreachable",
			Retryable: true,
		})
		if !localAvailable {
			envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
				Action: string(mcpschema.ToolIndexRefresh),
				Label:  "Refresh local index",
				Reason: "Create the local SQLite index for local-first search.",
			})
		}
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolIndexStatus),
			Label:  "Retry remote index status",
			Reason: "Remote index health could not be confirmed.",
		})
		return envelope, nil
	}
	remoteStatus := remote.SchemaStatus()
	envelope.Data.RemoteAvailable = remoteStatus.RemoteAvailable
	if remoteStatus.Fresh {
		envelope.Data.Fresh = true
	}
	if !remoteStatus.LastRefreshAt.IsZero() {
		envelope.Data.LastRefreshAt = remoteStatus.LastRefreshAt
	}
	envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
		Action: string(mcpschema.ToolSearchContext),
		Label:  "Search context",
		Reason: "Remote indexer is available.",
	})
	return envelope, nil
}

func (l *Local) IndexRefresh(ctx context.Context, req mcpschema.IndexRefreshRequest) (mcpschema.Envelope[mcpschema.IndexRefreshData], error) {
	report, err := l.Index.IndexWorkspace(ctx, l.Root)
	if err != nil {
		return mcpschema.Envelope[mcpschema.IndexRefreshData]{}, err
	}
	envelope := mcpschema.Envelope[mcpschema.IndexRefreshData]{
		OK: true,
		Data: mcpschema.IndexRefreshData{
			LocalRefreshed: true,
			MutationResult: mcpschema.MutationResult{ChangedPaths: []mcpschema.ChangedPath{{
				Path: ".cairn/index/cairn.db",
				Kind: "updated",
			}}},
			Message: fmt.Sprintf("indexed %d managed documents", len(report.Indexed)),
		},
		Provenance: l.provenance("index_refresh"),
	}
	if l.RemoteIndex == nil {
		envelope.NextSteps = []mcpschema.NextStep{{
			Action: string(mcpschema.ToolSearchContext),
			Label:  "Search local context",
			Reason: "Local metadata index refresh completed.",
		}}
		return envelope, nil
	}
	response, err := l.RemoteIndex.Refresh(ctx, remoteindex.RefreshRequest{
		WorkspaceID: l.provenance("index_refresh").WorkspaceID,
		DryRun:      req.DryRun,
	})
	if err != nil {
		envelope.OK = false
		envelope.Warnings = []mcpschema.Warning{{
			Code:    mcpschema.WarningRemoteService,
			Message: "remote index refresh failed: " + err.Error(),
		}}
		envelope.Unavailable = []mcpschema.UnavailableMode{{
			Mode:      "remote",
			Reason:    mcpschema.WarningRemoteService,
			Message:   "remote index refresh failed",
			Retryable: true,
		}}
		envelope.NextSteps = []mcpschema.NextStep{
			{
				Action: string(mcpschema.ToolSearchContext),
				Label:  "Use local search",
				Reason: "Local index refresh completed before the remote refresh failed.",
			},
			{
				Action: string(mcpschema.ToolIndexStatus),
				Label:  "Check index availability",
				Reason: "Inspect remote index health before retrying refresh.",
			},
		}
		return envelope, nil
	}
	envelope.Data.RemoteRefreshed = response.Refreshed
	envelope.Data.Accepted = response.Accepted
	envelope.Data.JobID = response.JobID
	envelope.Data.LastRefreshAt = response.LastRefreshAt
	if response.Message != "" {
		envelope.Data.Message = response.Message
	}
	if response.Accepted || response.Refreshed {
		envelope.Data.ChangedPaths = append(envelope.Data.ChangedPaths, mcpschema.ChangedPath{
			Path: ".cairn/index/remote",
			Kind: "refreshed",
		})
	}
	if response.Accepted && !response.Refreshed {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolIndexStatus),
			Label:  "Check refresh status",
			Reason: "The remote indexer accepted the refresh asynchronously.",
		})
	}
	if response.Refreshed {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolSearchContext),
			Label:  "Search refreshed context",
			Reason: "Local index refresh completed.",
		})
	}
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
