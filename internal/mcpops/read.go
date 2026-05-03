package mcpops

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
)

func (l *Local) ReadDocument(ctx context.Context, req mcpschema.ReadDocumentRequest) (mcpschema.Envelope[mcpschema.ReadDocumentData], error) {
	ref, err := l.resolveDocument(ctx, req.DocumentRef)
	if err != nil {
		return mcpschema.Envelope[mcpschema.ReadDocumentData]{}, err
	}
	content, err := os.ReadFile(filepath.Join(l.Root, filepath.FromSlash(ref.Path)))
	if err != nil {
		return mcpschema.Envelope[mcpschema.ReadDocumentData]{}, err
	}
	parsed, err := document.ParseMarkdown(string(content))
	if err != nil {
		return mcpschema.Envelope[mcpschema.ReadDocumentData]{}, err
	}
	if !parsed.HasFrontmatter {
		return mcpschema.Envelope[mcpschema.ReadDocumentData]{}, fmt.Errorf("read_document requires managed document frontmatter")
	}

	body := documentBody(string(content), parsed.ContentStartLine)
	sections := parseSections(body)
	mode := req.Mode
	if mode == "" {
		mode = mcpschema.ReadModeSummary
	}
	data := mcpschema.ReadDocumentData{
		Document: summaryFromMetadata(ref.Path, parsed.Metadata),
		Mode:     mode,
	}
	envelope := mcpschema.Envelope[mcpschema.ReadDocumentData]{
		OK:         true,
		Provenance: l.provenance("document"),
	}

	switch mode {
	case mcpschema.ReadModeFrontmatter:
		data.Frontmatter = parsed.Frontmatter
	case mcpschema.ReadModeTOC:
		data.TOC = headings(sections)
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolReadDocument),
			Label:  "Read specific sections",
			Reason: "Use sections mode with headings from the table of contents.",
		})
	case mcpschema.ReadModeSections:
		selected, missing := selectSections(sections, req.Sections)
		data.Sections = selected
		for _, heading := range missing {
			envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
				Code:       mcpschema.WarningProgressiveRead,
				Message:    "requested section was not found",
				Path:       ref.Path,
				DocumentID: parsed.Metadata.ID,
				Details:    map[string]string{"heading": heading},
			})
		}
	case mcpschema.ReadModeFull:
		data.Content = body
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolReadDocument),
			Label:  "Use summary or sections next time",
			Reason: "Full reads are available but should not be the default.",
		})
	default:
		data.Mode = mcpschema.ReadModeSummary
		data.Summary = summarize(body)
		data.TOC = headings(sections)
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolReadDocument),
			Label:  "Inspect table of contents",
			Reason: "Use toc or sections mode for progressive disclosure.",
		})
	}

	envelope.Data = data
	return envelope, nil
}

func (l *Local) resolveDocument(ctx context.Context, ref mcpschema.DocumentRef) (mcpschema.DocumentSummary, error) {
	query := localindex.Query{
		ID:    ref.ID,
		Slug:  ref.Slug,
		Path:  ref.Path,
		Limit: 1,
	}
	results, err := l.Index.Query(ctx, query)
	if err != nil {
		return mcpschema.DocumentSummary{}, err
	}
	if len(results) == 0 {
		return mcpschema.DocumentSummary{}, fmt.Errorf("document not found")
	}
	return summaries(results)[0], nil
}

type parsedSection struct {
	Heading string
	Level   int
	Content string
}

func parseSections(body string) []parsedSection {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	var sections []parsedSection
	current := parsedSection{Heading: "", Level: 0}
	var content []string
	flush := func() {
		current.Content = strings.TrimSpace(strings.Join(content, "\n"))
		if current.Heading != "" || current.Content != "" {
			sections = append(sections, current)
		}
		content = nil
	}
	for _, line := range lines {
		level, heading, ok := parseHeading(line)
		if ok {
			flush()
			current = parsedSection{Heading: heading, Level: level}
			continue
		}
		content = append(content, line)
	}
	flush()
	return sections
}

func parseHeading(line string) (int, string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return 0, "", false
	}
	level := 0
	for level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(trimmed) || trimmed[level] != ' ' {
		return 0, "", false
	}
	heading := strings.TrimSpace(trimmed[level:])
	if heading == "" {
		return 0, "", false
	}
	return level, heading, true
}

func headings(sections []parsedSection) []mcpschema.Heading {
	var result []mcpschema.Heading
	for _, section := range sections {
		if section.Heading == "" {
			continue
		}
		result = append(result, mcpschema.Heading{
			Level: section.Level,
			Text:  section.Heading,
			ID:    headingID(section.Heading),
		})
	}
	return result
}

func selectSections(sections []parsedSection, requested []string) ([]mcpschema.DocumentSection, []string) {
	byHeading := map[string]parsedSection{}
	for _, section := range sections {
		byHeading[strings.ToLower(section.Heading)] = section
	}
	var selected []mcpschema.DocumentSection
	var missing []string
	for _, heading := range requested {
		section, ok := byHeading[strings.ToLower(strings.TrimSpace(heading))]
		if !ok {
			missing = append(missing, heading)
			continue
		}
		selected = append(selected, mcpschema.DocumentSection{
			Heading: section.Heading,
			Content: section.Content,
		})
	}
	return selected, missing
}

func summaryFromMetadata(path string, metadata document.Metadata) mcpschema.DocumentSummary {
	return mcpschema.DocumentSummary{
		ID:      metadata.ID,
		Path:    path,
		Title:   metadata.Title,
		Slug:    metadata.Slug,
		Type:    metadata.Type,
		Status:  metadata.Status,
		Tags:    metadata.Tags,
		Updated: metadata.Updated,
		Authors: metadata.Authors,
		Actors:  metadata.Actors,
		Source:  metadata.Source,
	}
}

func summarize(body string) string {
	cleaned := strings.TrimSpace(body)
	if len(cleaned) <= 240 {
		return cleaned
	}
	return strings.TrimSpace(cleaned[:240]) + "..."
}

func headingID(heading string) string {
	lower := strings.ToLower(heading)
	var b strings.Builder
	lastDash := false
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func documentBody(content string, contentStartLine int) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	start := contentStartLine - 1
	if start < 0 || start > len(lines) {
		start = 0
	}
	return strings.Join(lines[start:], "\n")
}
