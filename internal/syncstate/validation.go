package syncstate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/document"
)

type ValidationFinding struct {
	Path    string
	Message string
}

type ValidationError struct {
	Findings []ValidationFinding
}

func (e ValidationError) Error() string {
	if len(e.Findings) == 0 {
		return "sync validation failed"
	}
	return "sync validation failed: " + e.Findings[0].Path + ": " + e.Findings[0].Message
}

func validateLocalManifest(root string, manifest Manifest) error {
	cfg, _ := document.LoadConfig(root)
	var findings []ValidationFinding
	for _, entry := range manifest.Entries {
		if !strings.EqualFold(filepath.Ext(entry.Path), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(entry.Path)))
		if err != nil {
			return err
		}
		findings = append(findings, validateMarkdownContent(entry.Path, content, cfg)...)
	}
	if len(findings) > 0 {
		return ValidationError{Findings: findings}
	}
	return nil
}

func validateRemoteContent(root string, path string, content []byte) error {
	if !strings.EqualFold(filepath.Ext(path), ".md") {
		return nil
	}
	cfg, _ := document.LoadConfig(root)
	findings := validateMarkdownContent(path, content, cfg)
	if len(findings) > 0 {
		return ValidationError{Findings: findings}
	}
	return nil
}

func validateMarkdownContent(path string, content []byte, cfg document.Config) []ValidationFinding {
	parsed, err := document.ParseMarkdown(string(content))
	if err != nil {
		return []ValidationFinding{{Path: path, Message: err.Error()}}
	}
	if !isSyncManagedMarkdown(path, parsed, cfg) {
		return nil
	}
	result := document.Validate(parsed, document.ValidationModeDurableBoundary)
	if !result.Blocking() {
		return nil
	}
	findings := make([]ValidationFinding, 0, len(result.Findings))
	for _, finding := range result.Findings {
		if finding.Severity != document.SeverityError {
			continue
		}
		message := finding.Message
		if finding.Field != "" {
			message = fmt.Sprintf("%s: %s", finding.Field, finding.Message)
		}
		findings = append(findings, ValidationFinding{Path: path, Message: message})
	}
	return findings
}

func isSyncManagedMarkdown(path string, parsed document.ParseResult, cfg document.Config) bool {
	if parsed.HasFrontmatter && parsed.Metadata.ID != "" {
		return true
	}
	for folder := range cfg.ManagedFolderSet() {
		if path == folder || strings.HasPrefix(path, folder+"/") {
			return true
		}
	}
	return false
}
