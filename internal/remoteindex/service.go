package remoteindex

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cairn/internal/document"
	"cairn/internal/mcpschema"
)

type Service struct {
	Root      string
	Now       func() time.Time
	refreshed time.Time
	count     int
}

func NewService(root string) *Service {
	return &Service{Root: root}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/index/status", s.handleStatus)
	mux.HandleFunc("/index/refresh", s.handleRefresh)
	mux.HandleFunc("/search", s.handleSearch)
	return mux
}

func (s *Service) handleStatus(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, StatusResponse{
		Available:     true,
		Fresh:         !s.refreshed.IsZero(),
		LastRefreshAt: s.refreshed,
		IndexedCount:  s.count,
		Message:       "local prototype indexer is available",
	})
}

func (s *Service) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}
	var req RefreshRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	count, err := countManagedMarkdown(s.Root)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	if !req.DryRun {
		s.refreshed = now
		s.count = count
	}
	writeJSON(w, http.StatusOK, RefreshResponse{
		Accepted:      true,
		Refreshed:     !req.DryRun,
		LastRefreshAt: now,
		Message:       "local prototype refresh completed without embeddings",
	})
}

func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}
	var req SearchRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 25
	}
	results, err := searchManagedMarkdown(s.Root, req.Query, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, SearchResponse{
		Results:        results,
		AttemptedModes: []mcpschema.SearchMode{mcpschema.SearchModeSemantic},
		Provenance: mcpschema.ItemProvenance{
			Source: "local_cocoindex_prototype",
		},
	})
}

func countManagedMarkdown(root string) (int, error) {
	results, err := searchManagedMarkdown(root, "", 0)
	return len(results), err
}

func searchManagedMarkdown(root string, query string, limit int) ([]SearchResult, error) {
	needle := strings.ToLower(strings.TrimSpace(query))
	var results []SearchResult
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			if rel == ".git" || rel == ".cairn" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(rel), ".md") {
			return nil
		}
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(contentBytes)
		parsed, err := document.ParseMarkdown(content)
		if err != nil || !parsed.HasFrontmatter || parsed.Metadata.ID == "" {
			return nil
		}
		if needle != "" && !strings.Contains(strings.ToLower(content), needle) {
			return nil
		}
		results = append(results, SearchResult{
			Path:    rel,
			Title:   parsed.Metadata.Title,
			Type:    parsed.Metadata.Type,
			Status:  parsed.Metadata.Status,
			Slug:    parsed.Metadata.Slug,
			Tags:    parsed.Metadata.Tags,
			Updated: parsed.Metadata.Updated,
			Score:   score(content, needle),
			Snippet: searchSnippet(content, needle),
			Authors: parsed.Metadata.Authors,
			Actors:  parsed.Metadata.Actors,
			Source:  parsed.Metadata.Source,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Path < results[j].Path
		}
		return results[i].Score > results[j].Score
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func score(content string, needle string) float64 {
	if needle == "" {
		return 1
	}
	return float64(strings.Count(strings.ToLower(content), needle))
}

func searchSnippet(content string, needle string) string {
	if needle == "" {
		return ""
	}
	index := strings.Index(strings.ToLower(content), needle)
	if index < 0 {
		return ""
	}
	start := index - 40
	if start < 0 {
		start = 0
	}
	end := index + len(needle) + 80
	if end > len(content) {
		end = len(content)
	}
	return strings.TrimSpace(content[start:end])
}

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	w.Header().Set("Allow", http.MethodPost)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	return false
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
