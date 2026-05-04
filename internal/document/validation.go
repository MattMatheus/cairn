package document

import (
	"fmt"
	"regexp"
	"sort"
	"time"
)

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

type ValidationMode string

const (
	ValidationModeDiscovery       ValidationMode = "discovery"
	ValidationModeDurableBoundary ValidationMode = "durable_boundary"
)

type Finding struct {
	Severity Severity
	Field    string
	Message  string
}

func (f Finding) Blocking() bool {
	return f.Severity == SeverityError
}

type ValidationResult struct {
	Findings []Finding
}

func (r ValidationResult) Blocking() bool {
	for _, finding := range r.Findings {
		if finding.Blocking() {
			return true
		}
	}
	return false
}

func Validate(result ParseResult, mode ValidationMode) ValidationResult {
	var findings []Finding
	blockingSeverity := SeverityWarning
	if mode == ValidationModeDurableBoundary {
		blockingSeverity = SeverityError
	}

	if !result.HasFrontmatter {
		return ValidationResult{Findings: []Finding{{
			Severity: blockingSeverity,
			Field:    "frontmatter",
			Message:  "missing frontmatter",
		}}}
	}

	for _, warning := range result.ParseWarnings {
		findings = append(findings, Finding{
			Severity: SeverityWarning,
			Field:    "frontmatter",
			Message:  warning,
		})
	}

	required := []string{"id", "schema_version", "title", "slug", "type", "status", "created", "updated", "authors", "actors", "source", "tags"}
	for _, field := range required {
		if _, ok := result.Frontmatter[field]; !ok {
			findings = append(findings, Finding{
				Severity: blockingSeverity,
				Field:    field,
				Message:  fmt.Sprintf("missing required field %q", field),
			})
		}
	}

	metadata := result.Metadata
	validateString(&findings, blockingSeverity, result.Frontmatter, "id", metadata.ID, validID)
	validateSchemaVersion(&findings, blockingSeverity, result.Frontmatter, metadata.SchemaVersion)
	validateString(&findings, blockingSeverity, result.Frontmatter, "title", metadata.Title, nonEmpty)
	validateString(&findings, blockingSeverity, result.Frontmatter, "slug", metadata.Slug, validSlug)
	validateStringSet(&findings, blockingSeverity, result.Frontmatter, "type", metadata.Type, validDocumentTypes)
	validateStringSet(&findings, blockingSeverity, result.Frontmatter, "status", metadata.Status, validStatuses)
	validateTime(&findings, blockingSeverity, result.Frontmatter, "created", metadata.Created)
	validateTime(&findings, blockingSeverity, result.Frontmatter, "updated", metadata.Updated)
	validateStringSlice(&findings, blockingSeverity, result.Frontmatter, "authors", metadata.Authors)
	validateStringSlice(&findings, blockingSeverity, result.Frontmatter, "actors", metadata.Actors)
	validateString(&findings, blockingSeverity, result.Frontmatter, "source", metadata.Source, nonEmpty)
	validateTags(&findings, blockingSeverity, result.Frontmatter, metadata.Tags)

	sort.Strings(metadata.UnknownFields)
	for _, field := range metadata.UnknownFields {
		findings = append(findings, Finding{
			Severity: SeverityWarning,
			Field:    field,
			Message:  fmt.Sprintf("unknown frontmatter field %q", field),
		})
	}

	return ValidationResult{Findings: findings}
}

var (
	validID   = regexp.MustCompile(`^cairn:[A-Za-z0-9]+$`).MatchString
	validSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString
	nonEmpty  = regexp.MustCompile(`\S`).MatchString
)

var coreFields = map[string]struct{}{
	"id":             {},
	"schema_version": {},
	"title":          {},
	"slug":           {},
	"type":           {},
	"status":         {},
	"created":        {},
	"updated":        {},
	"authors":        {},
	"actors":         {},
	"source":         {},
	"tags":           {},
}

var validStatuses = map[string]struct{}{
	"inbox":     {},
	"draft":     {},
	"working":   {},
	"proposed":  {},
	"canonical": {},
	"archived":  {},
}

var validDocumentTypes = map[string]struct{}{
	"note":          {},
	"handoff":       {},
	"investigation": {},
	"decision":      {},
	"runbook":       {},
	"project":       {},
	"service":       {},
	"onboarding":    {},
}

func validateString(findings *[]Finding, severity Severity, fields map[string]any, field string, value string, valid func(string) bool) {
	if !fieldExists(fields, field) {
		return
	}
	if _, ok := fields[field].(string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("%s must be a string", field),
		})
		return
	}
	if !valid(value) {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("invalid %s", field),
		})
	}
}

func validateSchemaVersion(findings *[]Finding, severity Severity, fields map[string]any, value int) {
	if !fieldExists(fields, "schema_version") {
		return
	}
	if _, ok := fields["schema_version"].(string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    "schema_version",
			Message:  "schema_version must be an integer",
		})
		return
	}
	if value < 1 {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    "schema_version",
			Message:  "schema_version must be positive",
		})
	}
}

func validateStringSet(findings *[]Finding, severity Severity, fields map[string]any, field string, value string, allowed map[string]struct{}) {
	if !fieldExists(fields, field) {
		return
	}
	if _, ok := fields[field].(string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("%s must be a string", field),
		})
		return
	}
	if _, ok := allowed[value]; !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("unknown %s %q", field, value),
		})
	}
}

func validateTime(findings *[]Finding, severity Severity, fields map[string]any, field string, value time.Time) {
	if !fieldExists(fields, field) {
		return
	}
	if _, ok := fields[field].(string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("%s must be a timestamp string", field),
		})
		return
	}
	if value.IsZero() {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("missing or invalid timestamp %q", field),
		})
	}
}

func validateStringSlice(findings *[]Finding, severity Severity, fields map[string]any, field string, values []string) {
	if !fieldExists(fields, field) {
		return
	}
	if _, ok := fields[field].([]string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    field,
			Message:  fmt.Sprintf("%s must be a string array", field),
		})
		return
	}
	for _, value := range values {
		if value == "" {
			*findings = append(*findings, Finding{
				Severity: severity,
				Field:    field,
				Message:  fmt.Sprintf("%s contains an empty value", field),
			})
		}
	}
}

func validateTags(findings *[]Finding, severity Severity, fields map[string]any, values []string) {
	if !fieldExists(fields, "tags") {
		return
	}
	if _, ok := fields["tags"].([]string); !ok {
		*findings = append(*findings, Finding{
			Severity: severity,
			Field:    "tags",
			Message:  "tags must be a string array",
		})
		return
	}
	for _, value := range values {
		if value == "" {
			*findings = append(*findings, Finding{
				Severity: severity,
				Field:    "tags",
				Message:  "tags contains an empty value",
			})
			continue
		}
		if value[0] == '#' {
			*findings = append(*findings, Finding{
				Severity: severity,
				Field:    "tags",
				Message:  "tags should not include inline # prefix",
			})
		}
	}
}

func fieldExists(fields map[string]any, field string) bool {
	_, ok := fields[field]
	return ok
}
