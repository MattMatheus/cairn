package document

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Workspace struct {
	Root  string
	Now   func() time.Time
	NewID func() string
}

type OperationResult struct {
	Path         string
	OriginalPath string
	DocumentID   string
	NextSteps    []string
}

type CaptureOptions struct {
	Actor   string
	Title   string
	Body    string
	Type    string
	Authors []string
	Tags    []string
}

type PromoteOptions struct {
	Path   string
	Type   string
	Status string
}

type ArchiveOptions struct {
	Path string
}

type PurgeOptions struct {
	Path string
}

func (w Workspace) Capture(opts CaptureOptions) (OperationResult, error) {
	if opts.Actor == "" {
		return OperationResult{}, errors.New("actor is required")
	}
	if err := validatePathSegment("actor", opts.Actor); err != nil {
		return OperationResult{}, err
	}
	if opts.Title == "" {
		return OperationResult{}, errors.New("title is required")
	}

	now := w.now()
	docType := defaultString(opts.Type, "note")
	slug := slugify(opts.Title)
	if slug == "" {
		return OperationResult{}, errors.New("title must produce a non-empty slug")
	}
	authors := opts.Authors
	if len(authors) == 0 {
		authors = []string{opts.Actor}
	}

	metadata := Metadata{
		ID:            w.newID(),
		SchemaVersion: 1,
		Title:         opts.Title,
		Slug:          slug,
		Type:          docType,
		Status:        "working",
		Created:       now,
		Updated:       now,
		Authors:       authors,
		Actors:        []string{opts.Actor},
		Source:        "capture",
		Tags:          opts.Tags,
	}

	rendered := renderDocument(metadata, opts.Body)
	parsed, err := ParseMarkdown(rendered)
	if err != nil {
		return OperationResult{}, err
	}
	if validation := Validate(parsed, ValidationModeDurableBoundary); validation.Blocking() {
		return OperationResult{}, fmt.Errorf("capture would create invalid frontmatter: %s", formatFindings(validation.Findings))
	}

	relativePath := filepath.Join("agents", opts.Actor, slug+".md")
	absolutePath := w.absolute(relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return OperationResult{}, err
	}
	if err := os.WriteFile(absolutePath, []byte(rendered), 0o644); err != nil {
		return OperationResult{}, err
	}

	return OperationResult{
		Path:       relativePath,
		DocumentID: metadata.ID,
		NextSteps:  []string{"promote the document when it is ready for review", "sync the workspace when sharing is needed"},
	}, nil
}

func (w Workspace) Promote(opts PromoteOptions) (OperationResult, error) {
	if opts.Path == "" {
		return OperationResult{}, errors.New("path is required")
	}
	sourcePath, err := cleanWorkspacePath(opts.Path)
	if err != nil {
		return OperationResult{}, err
	}
	targetStatus := defaultString(opts.Status, "proposed")
	if targetStatus != "proposed" && targetStatus != "canonical" {
		return OperationResult{}, fmt.Errorf("unsupported promotion status %q", targetStatus)
	}

	absolutePath := w.absolute(sourcePath)
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return OperationResult{}, err
	}

	parsed, err := ParseMarkdown(string(content))
	if err != nil {
		return OperationResult{}, err
	}
	body := bodyFromMarkdown(string(content), parsed)

	if targetStatus == "canonical" {
		validation := Validate(parsed, ValidationModeDurableBoundary)
		if validation.Blocking() {
			return OperationResult{}, fmt.Errorf("canonical promotion blocked by invalid frontmatter")
		}
	}

	metadata := parsed.Metadata
	if targetStatus == "proposed" {
		metadata = w.repairMetadata(metadata, opts, sourcePath)
	} else {
		metadata.Type = defaultString(opts.Type, metadata.Type)
		metadata.Status = "canonical"
		metadata.Updated = w.now()
	}

	if targetStatus == "proposed" {
		metadata.Status = "proposed"
		metadata.Updated = w.now()
	}

	targetPath, err := w.promotionTargetPath(metadata, targetStatus)
	if err != nil {
		return OperationResult{}, err
	}
	targetAbsolute := w.absolute(targetPath)
	if err := os.MkdirAll(filepath.Dir(targetAbsolute), 0o755); err != nil {
		return OperationResult{}, err
	}
	if err := os.WriteFile(targetAbsolute, []byte(renderDocument(metadata, body)), 0o644); err != nil {
		return OperationResult{}, err
	}
	if filepath.Clean(targetAbsolute) != filepath.Clean(absolutePath) {
		if err := os.Remove(absolutePath); err != nil {
			return OperationResult{}, err
		}
	}

	return OperationResult{
		Path:         targetPath,
		OriginalPath: sourcePath,
		DocumentID:   metadata.ID,
		NextSteps:    []string{"review the promoted document", "sync the workspace when sharing is needed"},
	}, nil
}

func (w Workspace) Archive(opts ArchiveOptions) (OperationResult, error) {
	if opts.Path == "" {
		return OperationResult{}, errors.New("path is required")
	}
	sourcePath, err := cleanWorkspacePath(opts.Path)
	if err != nil {
		return OperationResult{}, err
	}

	absolutePath := w.absolute(sourcePath)
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return OperationResult{}, err
	}
	parsed, err := ParseMarkdown(string(content))
	if err != nil {
		return OperationResult{}, err
	}
	if !parsed.HasFrontmatter {
		return OperationResult{}, errors.New("archive requires managed document frontmatter")
	}

	metadata := parsed.Metadata
	metadata.Status = "archived"
	metadata.Updated = w.now()
	body := bodyFromMarkdown(string(content), parsed)
	targetPath := filepath.Join("archive", sourcePath)
	targetAbsolute := w.absolute(targetPath)
	if err := os.MkdirAll(filepath.Dir(targetAbsolute), 0o755); err != nil {
		return OperationResult{}, err
	}
	if err := os.WriteFile(targetAbsolute, []byte(renderDocument(metadata, body)), 0o644); err != nil {
		return OperationResult{}, err
	}
	if err := os.Remove(absolutePath); err != nil {
		return OperationResult{}, err
	}

	return OperationResult{
		Path:         targetPath,
		OriginalPath: sourcePath,
		DocumentID:   metadata.ID,
		NextSteps:    []string{"keep archived content for history", "use CLI purge only when hard deletion is required"},
	}, nil
}

func (w Workspace) Purge(opts PurgeOptions) (OperationResult, error) {
	if opts.Path == "" {
		return OperationResult{}, errors.New("path is required")
	}
	sourcePath, err := cleanWorkspacePath(opts.Path)
	if err != nil {
		return OperationResult{}, err
	}
	if !isArchivePath(sourcePath) {
		return OperationResult{}, errors.New("purge requires a path under archive/")
	}

	absolutePath := w.absolute(sourcePath)
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return OperationResult{}, err
	}
	parsed, err := ParseMarkdown(string(content))
	if err != nil {
		return OperationResult{}, err
	}
	if !parsed.HasFrontmatter {
		return OperationResult{}, errors.New("purge requires managed document frontmatter")
	}
	if parsed.Metadata.Status != "archived" {
		return OperationResult{}, fmt.Errorf("purge requires archived status, got %q", parsed.Metadata.Status)
	}
	if err := os.Remove(absolutePath); err != nil {
		return OperationResult{}, err
	}

	return OperationResult{
		Path:       sourcePath,
		DocumentID: parsed.Metadata.ID,
		NextSteps:  []string{"run `cairn sync push` when sharing the deletion is needed"},
	}, nil
}

func (w Workspace) repairMetadata(metadata Metadata, opts PromoteOptions, path string) Metadata {
	now := w.now()
	if metadata.ID == "" {
		metadata.ID = w.newID()
	}
	if metadata.SchemaVersion == 0 {
		metadata.SchemaVersion = 1
	}
	if metadata.Title == "" {
		metadata.Title = titleFromPath(path)
	}
	if metadata.Slug == "" {
		metadata.Slug = slugify(metadata.Title)
	}
	if metadata.Type == "" {
		metadata.Type = defaultString(opts.Type, "note")
	}
	if opts.Type != "" {
		metadata.Type = opts.Type
	}
	if metadata.Created.IsZero() {
		metadata.Created = now
	}
	if metadata.Updated.IsZero() {
		metadata.Updated = now
	}
	if len(metadata.Authors) == 0 {
		metadata.Authors = []string{"unknown"}
	}
	if metadata.Actors == nil {
		metadata.Actors = []string{}
	}
	if metadata.Source == "" {
		metadata.Source = "promotion"
	}
	if metadata.Tags == nil {
		metadata.Tags = []string{}
	}
	return metadata
}

func (w Workspace) promotionTargetPath(metadata Metadata, status string) (string, error) {
	cfg, err := LoadConfig(w.Root)
	if err != nil {
		return "", err
	}
	if status == "proposed" {
		return filepath.Join(cfg.DestinationFolder(metadata.Type), metadata.Slug+".md"), nil
	}
	if metadata.Type == "decision" {
		number, err := w.nextADRNumber()
		if err != nil {
			return "", err
		}
		return filepath.Join(cfg.DestinationFolder(metadata.Type), fmt.Sprintf("ADR-%04d-%s.md", number, metadata.Slug)), nil
	}
	return filepath.Join(cfg.DestinationFolder(metadata.Type), metadata.Slug+".md"), nil
}

func (w Workspace) nextADRNumber() (int, error) {
	entries, err := os.ReadDir(w.absolute("decisions"))
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	next := 1
	re := regexp.MustCompile(`^ADR-(\d{4,})-.*\.md$`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := re.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		value, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if value >= next {
			next = value + 1
		}
	}
	return next, nil
}

func (w Workspace) absolute(path string) string {
	return filepath.Join(w.Root, path)
}

func (w Workspace) now() time.Time {
	if w.Now != nil {
		return w.Now().UTC()
	}
	return time.Now().UTC()
}

func (w Workspace) newID() string {
	if w.NewID != nil {
		return w.NewID()
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		panic(err)
	}
	return "cairn:" + strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(random), "=")
}

func renderDocument(metadata Metadata, body string) string {
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("id: %s\n", metadata.ID))
	builder.WriteString(fmt.Sprintf("schema_version: %d\n", metadata.SchemaVersion))
	builder.WriteString(fmt.Sprintf("title: %s\n", metadata.Title))
	builder.WriteString(fmt.Sprintf("slug: %s\n", metadata.Slug))
	builder.WriteString(fmt.Sprintf("type: %s\n", metadata.Type))
	builder.WriteString(fmt.Sprintf("status: %s\n", metadata.Status))
	builder.WriteString(fmt.Sprintf("created: %s\n", metadata.Created.UTC().Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("updated: %s\n", metadata.Updated.UTC().Format(time.RFC3339)))
	writeStringArray(&builder, "authors", metadata.Authors)
	writeStringArray(&builder, "actors", metadata.Actors)
	builder.WriteString(fmt.Sprintf("source: %s\n", metadata.Source))
	writeStringArray(&builder, "tags", metadata.Tags)
	builder.WriteString("---\n\n")
	builder.WriteString(strings.TrimLeft(body, "\n"))
	return builder.String()
}

func writeStringArray(builder *strings.Builder, key string, values []string) {
	builder.WriteString(key + ":\n")
	for _, value := range values {
		builder.WriteString(fmt.Sprintf("  - %s\n", value))
	}
}

func bodyFromMarkdown(content string, parsed ParseResult) string {
	if !parsed.HasFrontmatter {
		return content
	}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	start := parsed.ContentStartLine - 1
	if start < 0 || start >= len(lines) {
		return ""
	}
	return strings.Join(lines[start:], "\n")
}

func slugify(value string) string {
	lower := strings.ToLower(value)
	var builder strings.Builder
	lastDash := false
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func titleFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	parts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func defaultString(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func cleanWorkspacePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("workspace path must be relative: %s", path)
	}
	clean := filepath.Clean(path)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("workspace path escapes root: %s", path)
	}
	return clean, nil
}

func isArchivePath(path string) bool {
	return path == "archive" || strings.HasPrefix(path, "archive"+string(filepath.Separator))
}

func validatePathSegment(name string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	if filepath.IsAbs(value) || strings.ContainsAny(value, `/\`) || value == "." || value == ".." {
		return fmt.Errorf("%s must be a single relative path segment", name)
	}
	return nil
}

func formatFindings(findings []Finding) string {
	messages := make([]string, 0, len(findings))
	for _, finding := range findings {
		if finding.Blocking() {
			messages = append(messages, finding.Field+": "+finding.Message)
		}
	}
	if len(messages) == 0 {
		return "unknown validation error"
	}
	return strings.Join(messages, "; ")
}
