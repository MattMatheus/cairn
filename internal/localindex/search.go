package localindex

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cairn/internal/document"
	"cairn/internal/mcpschema"
	"cairn/internal/remoteindex"
)

type SearchOptions struct {
	Query       string
	Mode        mcpschema.SearchMode
	Limit       int
	WorkspaceID string
	Remote      remoteindex.Client
}

func (i *Index) Search(ctx context.Context, root string, opts SearchOptions) (mcpschema.Envelope[mcpschema.SearchContextData], error) {
	mode := opts.Mode
	if mode == "" {
		mode = mcpschema.SearchModeAuto
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 25
	}

	envelope := mcpschema.Envelope[mcpschema.SearchContextData]{
		OK: true,
		Provenance: mcpschema.ResponseProvenance{
			Profile: mcpschema.ProfileLocal,
			Source:  "local_index",
		},
	}

	var results []mcpschema.SearchResult
	var attempted []mcpschema.SearchMode
	switch mode {
	case mcpschema.SearchModeMetadata:
		attempted = append(attempted, mcpschema.SearchModeMetadata)
		results, err := i.Query(ctx, Query{Text: opts.Query, Limit: limit})
		if err != nil {
			return envelope, err
		}
		envelope.Data = mcpschema.SearchContextData{Results: results, AttemptedModes: attempted}
	case mcpschema.SearchModeFullText:
		attempted = append(attempted, mcpschema.SearchModeFullText)
		fullTextResults, err := i.FullText(ctx, root, opts.Query, limit)
		if err != nil {
			return envelope, err
		}
		envelope.Data = mcpschema.SearchContextData{Results: fullTextResults, AttemptedModes: attempted}
	case mcpschema.SearchModeSemantic:
		attempted = append(attempted, mcpschema.SearchModeSemantic)
		remoteResults, remoteOK, err := remoteSemanticSearch(ctx, opts, limit)
		if err != nil {
			return envelope, err
		}
		if remoteOK {
			results = appendUniqueResults(results, remoteResults.Data.Results, limit)
			envelope.Warnings = append(envelope.Warnings, remoteResults.Warnings...)
			envelope.Unavailable = append(envelope.Unavailable, remoteResults.Unavailable...)
			envelope.NextSteps = append(envelope.NextSteps, remoteResults.NextSteps...)
		} else {
			envelope.Warnings = append(envelope.Warnings, unavailableSemanticWarning())
			envelope.Unavailable = append(envelope.Unavailable, unavailableSemanticMode())
			envelope.NextSteps = append(envelope.NextSteps, checkIndexStatusStep())
		}
		envelope.Data = mcpschema.SearchContextData{Results: results, AttemptedModes: attempted}
	default:
		attempted = append(attempted, mcpschema.SearchModeMetadata)
		metadataResults, err := i.Query(ctx, Query{Text: opts.Query, Limit: limit})
		if err != nil {
			return envelope, err
		}
		results = append(results, metadataResults...)

		attempted = append(attempted, mcpschema.SearchModeFullText)
		fullTextResults, err := i.FullText(ctx, root, opts.Query, limit)
		if err != nil {
			return envelope, err
		}
		results = appendUniqueResults(results, fullTextResults, limit)

		attempted = append(attempted, mcpschema.SearchModeSemantic)
		remoteResults, remoteOK, err := remoteSemanticSearch(ctx, opts, limit)
		if err != nil {
			return envelope, err
		}
		if remoteOK {
			results = appendUniqueResults(results, remoteResults.Data.Results, limit)
			envelope.Warnings = append(envelope.Warnings, remoteResults.Warnings...)
			envelope.Unavailable = append(envelope.Unavailable, remoteResults.Unavailable...)
			envelope.NextSteps = append(envelope.NextSteps, remoteResults.NextSteps...)
		} else {
			envelope.Warnings = append(envelope.Warnings, unavailableSemanticWarning(), unavailableRemoteWarning())
			envelope.Unavailable = append(envelope.Unavailable, unavailableSemanticMode(), unavailableRemoteMode())
			envelope.NextSteps = append(envelope.NextSteps, checkIndexStatusStep())
		}
		envelope.Data = mcpschema.SearchContextData{Results: results, AttemptedModes: attempted}
	}

	attemptedStrings := make([]string, 0, len(envelope.Data.AttemptedModes))
	for _, attemptedMode := range envelope.Data.AttemptedModes {
		attemptedStrings = append(attemptedStrings, string(attemptedMode))
	}
	envelope.Provenance.AttemptedModes = attemptedStrings
	return envelope, nil
}

func remoteSemanticSearch(ctx context.Context, opts SearchOptions, limit int) (mcpschema.Envelope[mcpschema.SearchContextData], bool, error) {
	if opts.Remote == nil {
		return mcpschema.Envelope[mcpschema.SearchContextData]{}, false, nil
	}
	response, err := opts.Remote.Search(ctx, remoteindex.SearchRequest{
		WorkspaceID: opts.WorkspaceID,
		Query:       opts.Query,
		Mode:        mcpschema.SearchModeSemantic,
		Limit:       limit,
	})
	if err != nil {
		envelope := mcpschema.Envelope[mcpschema.SearchContextData]{
			OK: false,
			Warnings: []mcpschema.Warning{{
				Code:    mcpschema.WarningRemoteService,
				Message: "remote indexer search failed: " + err.Error(),
			}},
			Unavailable: []mcpschema.UnavailableMode{{
				Mode:      "remote",
				Reason:    mcpschema.WarningRemoteService,
				Message:   "remote indexer search failed",
				Retryable: true,
			}},
			NextSteps: []mcpschema.NextStep{checkIndexStatusStep()},
		}
		return envelope, true, nil
	}
	return response.Envelope(), true, nil
}

func (i *Index) FullText(ctx context.Context, root string, query string, limit int) ([]mcpschema.SearchResult, error) {
	if limit <= 0 {
		limit = 25
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return nil, nil
	}

	var matches []fullTextMatch
	err := filepath.WalkDir(root, func(absolutePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if absolutePath == root {
			return nil
		}
		rel, err := filepath.Rel(root, absolutePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if rel == ".cairn" || rel == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(rel), ".md") {
			return nil
		}
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			return err
		}
		parsed, err := document.ParseMarkdown(string(content))
		if err != nil || !parsed.HasFrontmatter || parsed.Metadata.ID == "" {
			return nil
		}
		body := markdownBody(string(content), parsed.ContentStartLine)
		bodyLower := strings.ToLower(body)
		index := strings.Index(bodyLower, needle)
		if index < 0 {
			return nil
		}
		score := float64(strings.Count(bodyLower, needle))
		matches = append(matches, fullTextMatch{
			path:    rel,
			score:   score,
			snippet: snippet(body, index, len(query)),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].path < matches[j].path
		}
		return matches[i].score > matches[j].score
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}

	results := make([]mcpschema.SearchResult, 0, len(matches))
	for _, match := range matches {
		result, ok, err := i.Get(ctx, match.path)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result.MatchType = mcpschema.SearchModeFullText
		result.Score = match.score
		result.Snippet = match.snippet
		results = append(results, result)
	}
	return results, nil
}

type fullTextMatch struct {
	path    string
	score   float64
	snippet string
}

func appendUniqueResults(existing []mcpschema.SearchResult, next []mcpschema.SearchResult, limit int) []mcpschema.SearchResult {
	seen := map[string]bool{}
	for _, result := range existing {
		seen[result.Path] = true
	}
	for _, result := range next {
		if seen[result.Path] {
			continue
		}
		existing = append(existing, result)
		seen[result.Path] = true
		if len(existing) >= limit {
			return existing
		}
	}
	return existing
}

func markdownBody(content string, contentStartLine int) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	start := contentStartLine - 1
	if start < 0 || start > len(lines) {
		start = 0
	}
	return strings.Join(lines[start:], "\n")
}

func snippet(content string, index int, length int) string {
	start := index - 40
	if start < 0 {
		start = 0
	}
	end := index + length + 80
	if end > len(content) {
		end = len(content)
	}
	return strings.TrimSpace(content[start:end])
}

func unavailableSemanticWarning() mcpschema.Warning {
	return mcpschema.Warning{
		Code:    mcpschema.WarningIndexDegraded,
		Message: "semantic search is unavailable in local-only mode",
	}
}

func unavailableRemoteWarning() mcpschema.Warning {
	return mcpschema.Warning{
		Code:    mcpschema.WarningRemoteService,
		Message: "remote indexer is unavailable in local-only mode",
	}
}

func unavailableSemanticMode() mcpschema.UnavailableMode {
	return mcpschema.UnavailableMode{
		Mode:      string(mcpschema.SearchModeSemantic),
		Reason:    mcpschema.WarningIndexDegraded,
		Message:   "semantic index is not configured",
		Retryable: false,
	}
}

func unavailableRemoteMode() mcpschema.UnavailableMode {
	return mcpschema.UnavailableMode{
		Mode:      "remote",
		Reason:    mcpschema.WarningRemoteService,
		Message:   "remote indexer is not configured",
		Retryable: false,
	}
}

func checkIndexStatusStep() mcpschema.NextStep {
	return mcpschema.NextStep{
		Action: string(mcpschema.ToolIndexStatus),
		Label:  "Check index availability",
		Reason: "Use local results now or configure semantic/remote indexing later.",
	}
}
