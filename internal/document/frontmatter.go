package document

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Metadata struct {
	ID            string
	SchemaVersion int
	Title         string
	Slug          string
	Type          string
	Status        string
	Created       time.Time
	Updated       time.Time
	Authors       []string
	Actors        []string
	Source        string
	Tags          []string
	UnknownFields []string
}

type ParseResult struct {
	Metadata         Metadata
	HasFrontmatter   bool
	Frontmatter      map[string]any
	ParseWarnings    []string
	ContentStartLine int
}

func ParseMarkdown(input string) (ParseResult, error) {
	result := ParseResult{
		Frontmatter:      map[string]any{},
		ContentStartLine: 1,
	}

	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return result, nil
	}

	lines := strings.Split(normalized, "\n")
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return result, fmt.Errorf("frontmatter start marker found without closing marker")
	}

	result.HasFrontmatter = true
	result.ContentStartLine = end + 2
	fields, warnings, err := parseFrontmatterLines(lines[1:end])
	if err != nil {
		return result, err
	}
	result.Frontmatter = fields
	result.ParseWarnings = warnings
	result.Metadata = metadataFromFields(fields)
	return result, nil
}

func parseFrontmatterLines(lines []string) (map[string]any, []string, error) {
	fields := map[string]any{}
	var warnings []string

	for i := 0; i < len(lines); i++ {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			return fields, warnings, fmt.Errorf("line %d: list item without a field", i+1)
		}

		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			return fields, warnings, fmt.Errorf("line %d: expected key/value frontmatter field", i+1)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fields, warnings, fmt.Errorf("line %d: empty frontmatter key", i+1)
		}

		if value == "" {
			var values []string
			for i+1 < len(lines) {
				next := strings.TrimSpace(lines[i+1])
				if next == "" {
					i++
					continue
				}
				if !strings.HasPrefix(next, "- ") {
					break
				}
				values = append(values, unquote(strings.TrimSpace(strings.TrimPrefix(next, "- "))))
				i++
			}
			fields[key] = values
			continue
		}

		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			fields[key] = parseInlineArray(value)
			continue
		}

		fields[key] = unquote(value)
	}

	return fields, warnings, nil
}

func metadataFromFields(fields map[string]any) Metadata {
	var metadata Metadata

	metadata.ID = stringField(fields, "id")
	metadata.SchemaVersion = intField(fields, "schema_version")
	metadata.Title = stringField(fields, "title")
	metadata.Slug = stringField(fields, "slug")
	metadata.Type = stringField(fields, "type")
	metadata.Status = stringField(fields, "status")
	metadata.Created = timeField(fields, "created")
	metadata.Updated = timeField(fields, "updated")
	metadata.Authors = stringSliceField(fields, "authors")
	metadata.Actors = stringSliceField(fields, "actors")
	metadata.Source = stringField(fields, "source")
	metadata.Tags = stringSliceField(fields, "tags")

	for key := range fields {
		if _, ok := coreFields[key]; !ok {
			metadata.UnknownFields = append(metadata.UnknownFields, key)
		}
	}

	return metadata
}

func stringField(fields map[string]any, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func intField(fields map[string]any, key string) int {
	raw := stringField(fields, key)
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}

func timeField(fields map[string]any, key string) time.Time {
	raw := stringField(fields, key)
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return value
}

func stringSliceField(fields map[string]any, key string) []string {
	value, ok := fields[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return typed
	default:
		return nil
	}
}

func parseInlineArray(value string) []string {
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
	if inner == "" {
		return []string{}
	}
	parts := strings.Split(inner, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, unquote(strings.TrimSpace(part)))
	}
	return values
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}
	if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}
