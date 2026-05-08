package workspace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
	"cairn/internal/syncstate"
)

type HealthOptions struct {
	Now       func() time.Time
	StaleAge  time.Duration
	RecentMax int
}

type HealthReport struct {
	GeneratedAt        time.Time
	WorkspaceID        string
	TotalManaged       int
	CountsByType       map[string]int
	CountsByStatus     map[string]int
	Proposed           []mcpschema.DocumentSummary
	StaleWorking       []mcpschema.DocumentSummary
	RecentCanonical    []mcpschema.DocumentSummary
	ValidationFindings []mcpschema.ValidationFinding
	IndexAvailable     bool
	IndexFreshAt       time.Time
	SyncDiverged       bool
	LocalChangeCount   int
	RemoteChangeCount  int
	ConflictCount      int
}

func BuildHealthReport(ctx context.Context, root string, opts HealthOptions) (HealthReport, error) {
	now := time.Now().UTC()
	if opts.Now != nil {
		now = opts.Now().UTC()
	}
	staleAge := opts.StaleAge
	if staleAge <= 0 {
		staleAge = 30 * 24 * time.Hour
	}
	recentMax := opts.RecentMax
	if recentMax <= 0 {
		recentMax = 5
	}
	cfg, err := document.LoadConfig(root)
	if err != nil {
		return HealthReport{}, err
	}
	report := HealthReport{
		GeneratedAt:    now,
		WorkspaceID:    cfg.WorkspaceID,
		CountsByType:   map[string]int{},
		CountsByStatus: map[string]int{},
	}

	ignore, err := loadIgnore(root)
	if err != nil {
		return HealthReport{}, err
	}
	paths, err := markdownPaths(root, nil, ignore)
	if err != nil {
		return HealthReport{}, err
	}
	staleBefore := now.Add(-staleAge)
	for _, rel := range paths {
		if err := ctx.Err(); err != nil {
			return HealthReport{}, err
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return HealthReport{}, err
		}
		parsed, err := document.ParseMarkdown(string(content))
		if err != nil || !isManagedMarkdown(root, rel, parsed) || !parsed.HasFrontmatter || parsed.Metadata.ID == "" {
			continue
		}
		metadata := parsed.Metadata
		report.TotalManaged++
		report.CountsByType[metadata.Type]++
		report.CountsByStatus[metadata.Status]++
		summary := healthSummary(rel, metadata)
		switch metadata.Status {
		case "proposed":
			report.Proposed = append(report.Proposed, summary)
		case "working":
			if !metadata.Updated.IsZero() && metadata.Updated.Before(staleBefore) {
				report.StaleWorking = append(report.StaleWorking, summary)
			}
		case "canonical":
			report.RecentCanonical = append(report.RecentCanonical, summary)
		}
	}
	sortSummaries(report.Proposed)
	sortSummaries(report.StaleWorking)
	sort.Slice(report.RecentCanonical, func(i, j int) bool {
		if report.RecentCanonical[i].Updated.Equal(report.RecentCanonical[j].Updated) {
			return report.RecentCanonical[i].Path < report.RecentCanonical[j].Path
		}
		return report.RecentCanonical[i].Updated.After(report.RecentCanonical[j].Updated)
	})
	if len(report.RecentCanonical) > recentMax {
		report.RecentCanonical = report.RecentCanonical[:recentMax]
	}

	validation, err := Validate(ctx, root, ValidateOptions{Mode: document.ValidationModeDiscovery})
	if err != nil {
		return HealthReport{}, err
	}
	report.ValidationFindings = validation.Findings
	if info, err := os.Stat(localindex.DBPath(root)); err == nil {
		report.IndexAvailable = true
		report.IndexFreshAt = info.ModTime().UTC().Truncate(time.Second)
	}
	status, err := syncstate.StatusReport(ctx, root, syncstate.StatusOptions{WorkspaceID: cfg.WorkspaceID, Now: opts.Now})
	if err != nil {
		return HealthReport{}, err
	}
	report.SyncDiverged = status.Comparison.Diverged
	report.LocalChangeCount = len(status.Comparison.LocalChanges)
	report.RemoteChangeCount = len(status.Comparison.RemoteChanges)
	report.ConflictCount = len(status.Comparison.Conflicts)
	return report, nil
}

func RenderHealthReport(report HealthReport) string {
	var out bytes.Buffer
	fmt.Fprintln(&out, "# Cairn Knowledge Health")
	fmt.Fprintln(&out)
	fmt.Fprintf(&out, "- Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	if report.WorkspaceID != "" {
		fmt.Fprintf(&out, "- Workspace: %s\n", report.WorkspaceID)
	}
	fmt.Fprintf(&out, "- Managed documents: %d\n", report.TotalManaged)
	fmt.Fprintln(&out)
	writeCounts(&out, "Documents By Status", report.CountsByStatus)
	writeCounts(&out, "Documents By Type", report.CountsByType)
	writeSection(&out, "Proposed Documents Awaiting Review", report.Proposed)
	writeSection(&out, "Stale Working Documents", report.StaleWorking)
	writeSection(&out, "Recent Canonical Documents", report.RecentCanonical)
	fmt.Fprintln(&out, "## Validation Findings")
	if len(report.ValidationFindings) == 0 {
		fmt.Fprintln(&out)
		fmt.Fprintln(&out, "No validation findings.")
	} else {
		fmt.Fprintln(&out)
		for _, finding := range report.ValidationFindings {
			path := finding.Path
			if path == "" {
				path = "<workspace>"
			}
			fmt.Fprintf(&out, "- %s %s: %s\n", finding.Severity, path, finding.Message)
		}
	}
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "## Index And Sync")
	fmt.Fprintln(&out)
	if report.IndexAvailable {
		fmt.Fprintf(&out, "- Local index: available")
		if !report.IndexFreshAt.IsZero() {
			fmt.Fprintf(&out, " (last touched %s)", report.IndexFreshAt.Format(time.RFC3339))
		}
		fmt.Fprintln(&out)
	} else {
		fmt.Fprintln(&out, "- Local index: missing")
		fmt.Fprintln(&out, "- Next: run `cairn index refresh`.")
	}
	fmt.Fprintf(&out, "- Sync diverged: %t\n", report.SyncDiverged)
	fmt.Fprintf(&out, "- Local changes: %d\n", report.LocalChangeCount)
	fmt.Fprintf(&out, "- Remote changes: %d\n", report.RemoteChangeCount)
	fmt.Fprintf(&out, "- Conflicts: %d\n", report.ConflictCount)
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "## Suggested Follow-Up")
	fmt.Fprintln(&out)
	if len(report.ValidationFindings) > 0 {
		fmt.Fprintln(&out, "- Review validation findings before promoting or syncing durable knowledge.")
	}
	if len(report.Proposed) > 0 {
		fmt.Fprintln(&out, "- Review proposed documents and promote or archive them.")
	}
	if len(report.StaleWorking) > 0 {
		fmt.Fprintln(&out, "- Review stale working documents and decide whether to promote, refresh, or archive them.")
	}
	if !report.IndexAvailable {
		fmt.Fprintln(&out, "- Refresh the local index.")
	}
	if report.SyncDiverged || report.ConflictCount > 0 {
		fmt.Fprintln(&out, "- Resolve sync divergence before mutating shared state.")
	}
	if len(report.ValidationFindings) == 0 && len(report.Proposed) == 0 && len(report.StaleWorking) == 0 && report.IndexAvailable && !report.SyncDiverged && report.ConflictCount == 0 {
		fmt.Fprintln(&out, "- No immediate follow-up suggested.")
	}
	return out.String()
}

func healthSummary(path string, metadata document.Metadata) mcpschema.DocumentSummary {
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

func sortSummaries(values []mcpschema.DocumentSummary) {
	sort.Slice(values, func(i, j int) bool {
		return values[i].Path < values[j].Path
	})
}

func writeCounts(out *bytes.Buffer, title string, counts map[string]int) {
	fmt.Fprintf(out, "## %s\n\n", title)
	if len(counts) == 0 {
		fmt.Fprintln(out, "No managed documents.")
		fmt.Fprintln(out)
		return
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(out, "- %s: %d\n", key, counts[key])
	}
	fmt.Fprintln(out)
}

func writeSection(out *bytes.Buffer, title string, docs []mcpschema.DocumentSummary) {
	fmt.Fprintf(out, "## %s\n\n", title)
	if len(docs) == 0 {
		fmt.Fprintln(out, "None.")
		fmt.Fprintln(out)
		return
	}
	for _, doc := range docs {
		label := doc.Title
		if strings.TrimSpace(label) == "" {
			label = doc.Path
		}
		fmt.Fprintf(out, "- %s (`%s`)", label, doc.Path)
		if !doc.Updated.IsZero() {
			fmt.Fprintf(out, " updated %s", doc.Updated.Format("2006-01-02"))
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out)
}
