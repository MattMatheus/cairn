package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
	"cairn/internal/syncstate"
)

type ValidateOptions struct {
	Paths []string
	Mode  document.ValidationMode
}

func Validate(ctx context.Context, root string, opts ValidateOptions) (mcpschema.ValidateWorkspaceData, error) {
	if opts.Mode == "" {
		opts.Mode = document.ValidationModeDiscovery
	}

	ignore, err := loadIgnore(root)
	if err != nil {
		return mcpschema.ValidateWorkspaceData{}, err
	}

	var findings []mcpschema.ValidationFinding
	findings = append(findings, configFindings(root)...)
	paths, err := markdownPaths(root, opts.Paths, ignore)
	if err != nil {
		return mcpschema.ValidateWorkspaceData{}, err
	}
	for _, rel := range paths {
		if err := ctx.Err(); err != nil {
			return mcpschema.ValidateWorkspaceData{}, err
		}
		docFindings, err := validateDocument(root, rel, opts.Mode)
		if err != nil {
			return mcpschema.ValidateWorkspaceData{}, err
		}
		findings = append(findings, docFindings...)
	}

	findings = append(findings, metadataHealthFindings(root)...)
	return mcpschema.ValidateWorkspaceData{
		Findings: findings,
		Healthy:  !hasErrors(findings),
	}, nil
}

func configFindings(root string) []mcpschema.ValidationFinding {
	raw := document.ValidateConfigFiles(root)
	out := make([]mcpschema.ValidationFinding, 0, len(raw))
	for _, finding := range raw {
		out = append(out, mcpschema.ValidationFinding{
			Severity: finding.Severity,
			Code:     mcpschema.WarningValidation,
			Message:  finding.Message,
			Path:     finding.Path,
		})
	}
	return out
}

func validateDocument(root string, rel string, mode document.ValidationMode) ([]mcpschema.ValidationFinding, error) {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, err
	}
	parsed, err := document.ParseMarkdown(string(content))
	if err != nil {
		return []mcpschema.ValidationFinding{{
			Severity: severityForMode(mode),
			Code:     mcpschema.WarningValidation,
			Message:  err.Error(),
			Path:     rel,
		}}, nil
	}
	if !isManagedMarkdown(root, rel, parsed) {
		return nil, nil
	}

	result := document.Validate(parsed, mode)
	out := make([]mcpschema.ValidationFinding, 0, len(result.Findings))
	for _, finding := range result.Findings {
		message := finding.Message
		if finding.Field != "" {
			message = fmt.Sprintf("%s: %s", finding.Field, finding.Message)
		}
		out = append(out, mcpschema.ValidationFinding{
			Severity:   string(finding.Severity),
			Code:       mcpschema.WarningValidation,
			Message:    message,
			Path:       rel,
			DocumentID: parsed.Metadata.ID,
		})
	}
	return out, nil
}

func metadataHealthFindings(root string) []mcpschema.ValidationFinding {
	var findings []mcpschema.ValidationFinding
	if _, err := os.Stat(localindex.DBPath(root)); errors.Is(err, os.ErrNotExist) {
		// A missing local index is normal before the first `cairn index refresh`.
	} else if err != nil {
		findings = append(findings, mcpschema.ValidationFinding{
			Severity: "warning",
			Code:     mcpschema.WarningIndexDegraded,
			Message:  err.Error(),
			Path:     ".cairn/index/cairn.db",
		})
	}

	statePath := filepath.Join(root, ".cairn", "sync-state.json")
	if _, err := os.Stat(statePath); errors.Is(err, os.ErrNotExist) {
		// A missing sync state is normal before the first successful sync.
	} else if err != nil {
		findings = append(findings, mcpschema.ValidationFinding{
			Severity: "warning",
			Code:     mcpschema.WarningSyncDivergence,
			Message:  err.Error(),
			Path:     ".cairn/sync-state.json",
		})
	} else if _, err := syncstate.Load(root); err != nil {
		findings = append(findings, mcpschema.ValidationFinding{
			Severity: "error",
			Code:     mcpschema.WarningSyncDivergence,
			Message:  err.Error(),
			Path:     ".cairn/sync-state.json",
		})
	}

	return findings
}

func markdownPaths(root string, requested []string, ignore ignoreFile) ([]string, error) {
	if len(requested) > 0 {
		paths := make([]string, 0, len(requested))
		for _, raw := range requested {
			rel := cleanRel(raw)
			if rel == "" || strings.HasPrefix(rel, "../") || rel == ".." || ignore.matches(rel, false) || !strings.EqualFold(filepath.Ext(rel), ".md") {
				continue
			}
			paths = append(paths, rel)
		}
		return paths, nil
	}

	var paths []string
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
		rel = cleanRel(rel)
		if rel == ".git" || strings.HasPrefix(rel, ".git/") || rel == ".cairn" || strings.HasPrefix(rel, ".cairn/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if ignore.matches(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(rel), ".md") {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	return paths, err
}

func cleanRel(value string) string {
	rel := path.Clean(strings.TrimPrefix(filepath.ToSlash(value), "/"))
	if rel == "." {
		return ""
	}
	return rel
}

func severityForMode(mode document.ValidationMode) string {
	if mode == document.ValidationModeDurableBoundary {
		return string(document.SeverityError)
	}
	return string(document.SeverityWarning)
}

func hasErrors(findings []mcpschema.ValidationFinding) bool {
	for _, finding := range findings {
		if finding.Severity == string(document.SeverityError) {
			return true
		}
	}
	return false
}

func isManagedMarkdown(root string, rel string, parsed document.ParseResult) bool {
	if parsed.HasFrontmatter && parsed.Metadata.ID != "" {
		return true
	}
	cfg, err := document.LoadConfig(root)
	if err == nil {
		rel = filepath.ToSlash(rel)
		for folder := range cfg.ManagedFolderSet() {
			if rel == folder || strings.HasPrefix(rel, folder+"/") {
				return true
			}
		}
	}
	first, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
	switch first {
	case "inbox", "agents", "working", "decisions", "runbooks", "projects", "services", "handoffs", "onboarding", "archive":
		return true
	default:
		return false
	}
}
