package localindex

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cairn/internal/document"
	"cairn/internal/mcpschema"

	_ "modernc.org/sqlite"
)

const MatchTypeMetadata = mcpschema.SearchModeMetadata

type Index struct {
	db *sql.DB
}

type Metadata struct {
	Path    string
	Content string
}

type IndexReport struct {
	Indexed []string
	Skipped []SkippedFile
}

type SkippedFile struct {
	Path   string
	Reason string
}

type Query struct {
	Text   string
	Slug   string
	Tag    string
	Status string
	Type   string
	Path   string
	Actor  string
	Source string
	Recent bool
	Limit  int
}

func DBPath(root string) string {
	return filepath.Join(root, ".cairn", "index", "cairn.db")
}

func Open(root string) (*Index, error) {
	if err := os.MkdirAll(filepath.Dir(DBPath(root)), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", DBPath(root))
	if err != nil {
		return nil, err
	}
	index := &Index{db: db}
	if err := index.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return index, nil
}

func (i *Index) Close() error {
	if i == nil || i.db == nil {
		return nil
	}
	return i.db.Close()
}

func (i *Index) IndexWorkspace(ctx context.Context, root string) (IndexReport, error) {
	var report IndexReport
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
		indexed, reason, err := i.IndexMarkdownFile(ctx, root, rel)
		if err != nil {
			return err
		}
		if indexed {
			report.Indexed = append(report.Indexed, rel)
			return nil
		}
		report.Skipped = append(report.Skipped, SkippedFile{Path: rel, Reason: reason})
		return nil
	})
	return report, err
}

func (i *Index) IndexMarkdownFile(ctx context.Context, root string, relativePath string) (bool, string, error) {
	relativePath = cleanPath(relativePath)
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return false, "", err
	}
	parsed, err := document.ParseMarkdown(string(content))
	if err != nil {
		return false, "invalid_frontmatter", nil
	}
	if !parsed.HasFrontmatter || parsed.Metadata.ID == "" {
		return false, "unmanaged_markdown", nil
	}
	validation := document.Validate(parsed, document.ValidationModeDurableBoundary)
	if validation.Blocking() {
		return false, "invalid_metadata", nil
	}

	metadata := parsed.Metadata
	tags, err := marshalStrings(metadata.Tags)
	if err != nil {
		return false, "", err
	}
	authors, err := marshalStrings(metadata.Authors)
	if err != nil {
		return false, "", err
	}
	actors, err := marshalStrings(metadata.Actors)
	if err != nil {
		return false, "", err
	}

	_, err = i.db.ExecContext(ctx, `
		insert into documents (
			path, document_id, title, slug, type, status, tags, actors, authors, source, updated
		) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		on conflict(path) do update set
			document_id = excluded.document_id,
			title = excluded.title,
			slug = excluded.slug,
			type = excluded.type,
			status = excluded.status,
			tags = excluded.tags,
			actors = excluded.actors,
			authors = excluded.authors,
			source = excluded.source,
			updated = excluded.updated
	`, relativePath, metadata.ID, metadata.Title, metadata.Slug, metadata.Type, metadata.Status, tags, actors, authors, metadata.Source, formatTime(metadata.Updated))
	if err != nil {
		return false, "", err
	}
	return true, "", nil
}

func (i *Index) Query(ctx context.Context, query Query) ([]mcpschema.SearchResult, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 25
	}

	where := []string{"1 = 1"}
	args := []any{}
	if query.Text != "" {
		where = append(where, "(title like ? or slug like ? or path like ?)")
		like := "%" + query.Text + "%"
		args = append(args, like, like, like)
	}
	if query.Slug != "" {
		where = append(where, "slug = ?")
		args = append(args, query.Slug)
	}
	if query.Tag != "" {
		where = append(where, "tags like ?")
		args = append(args, "%\""+query.Tag+"\"%")
	}
	if query.Status != "" {
		where = append(where, "status = ?")
		args = append(args, query.Status)
	}
	if query.Type != "" {
		where = append(where, "type = ?")
		args = append(args, query.Type)
	}
	if query.Path != "" {
		where = append(where, "path = ?")
		args = append(args, cleanPath(query.Path))
	}
	if query.Actor != "" {
		where = append(where, "actors like ?")
		args = append(args, "%\""+query.Actor+"\"%")
	}
	if query.Source != "" {
		where = append(where, "source = ?")
		args = append(args, query.Source)
	}

	order := "title collate nocase asc, path asc"
	if query.Recent {
		order = "updated desc, path asc"
	}

	args = append(args, limit)
	rows, err := i.db.QueryContext(ctx, fmt.Sprintf(`
		select path, document_id, title, slug, type, status, tags, actors, authors, source, updated
		from documents
		where %s
		order by %s
		limit ?
	`, strings.Join(where, " and "), order), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []mcpschema.SearchResult
	for rows.Next() {
		var record documentRecord
		if err := rows.Scan(
			&record.Path,
			&record.DocumentID,
			&record.Title,
			&record.Slug,
			&record.Type,
			&record.Status,
			&record.TagsJSON,
			&record.ActorsJSON,
			&record.AuthorsJSON,
			&record.Source,
			&record.UpdatedRaw,
		); err != nil {
			return nil, err
		}
		result, err := record.searchResult()
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

func (i *Index) Get(ctx context.Context, path string) (mcpschema.SearchResult, bool, error) {
	results, err := i.Query(ctx, Query{Path: path, Limit: 1})
	if err != nil {
		return mcpschema.SearchResult{}, false, err
	}
	if len(results) == 0 {
		return mcpschema.SearchResult{}, false, nil
	}
	return results[0], true, nil
}

func (i *Index) migrate(ctx context.Context) error {
	if i == nil || i.db == nil {
		return errors.New("index is not open")
	}
	_, err := i.db.ExecContext(ctx, `
		create table if not exists documents (
			path text primary key,
			document_id text not null,
			title text not null,
			slug text not null,
			type text not null,
			status text not null,
			tags text not null,
			actors text not null,
			authors text not null,
			source text not null,
			updated text not null
		);
		create index if not exists idx_documents_title on documents(title);
		create index if not exists idx_documents_slug on documents(slug);
		create index if not exists idx_documents_status on documents(status);
		create index if not exists idx_documents_type on documents(type);
		create index if not exists idx_documents_source on documents(source);
		create index if not exists idx_documents_updated on documents(updated);
	`)
	return err
}

type documentRecord struct {
	Path        string
	DocumentID  string
	Title       string
	Slug        string
	Type        string
	Status      string
	TagsJSON    string
	ActorsJSON  string
	AuthorsJSON string
	Source      string
	UpdatedRaw  string
}

func (r documentRecord) searchResult() (mcpschema.SearchResult, error) {
	tags, err := unmarshalStrings(r.TagsJSON)
	if err != nil {
		return mcpschema.SearchResult{}, err
	}
	actors, err := unmarshalStrings(r.ActorsJSON)
	if err != nil {
		return mcpschema.SearchResult{}, err
	}
	authors, err := unmarshalStrings(r.AuthorsJSON)
	if err != nil {
		return mcpschema.SearchResult{}, err
	}
	updated, err := time.Parse(time.RFC3339, r.UpdatedRaw)
	if err != nil {
		return mcpschema.SearchResult{}, err
	}
	return mcpschema.SearchResult{
		Path:      r.Path,
		Title:     r.Title,
		Type:      r.Type,
		Status:    r.Status,
		Slug:      r.Slug,
		Tags:      tags,
		Updated:   updated,
		Score:     1,
		MatchType: MatchTypeMetadata,
		Snippet:   r.Title,
		Provenance: mcpschema.ItemProvenance{
			Authors: authors,
			Actors:  actors,
			Source:  r.Source,
		},
	}, nil
}

func marshalStrings(values []string) (string, error) {
	if values == nil {
		values = []string{}
	}
	content, err := json.Marshal(values)
	return string(content), err
}

func unmarshalStrings(value string) ([]string, error) {
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func cleanPath(value string) string {
	return strings.TrimPrefix(filepath.ToSlash(filepath.Clean(value)), "/")
}
