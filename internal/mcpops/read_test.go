package mcpops

import (
	"context"
	"testing"

	"cairn/internal/mcpschema"
)

func TestReadDocumentSupportsSummaryFrontmatterTOCSectionsAndFull(t *testing.T) {
	ops := newFixtureOps(t)
	writeFile(t, ops.Root, "runbooks/deploy.md", managedMarkdown("cairn:deploy", "Deploy Runbook", "deploy-runbook", "runbook", "canonical", []string{"deploy"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-04T12:00:00Z", "Intro paragraph.\n\n## Steps\n\n1. Build\n2. Ship\n\n## Rollback\n\nUndo the release.\n"))
	if _, err := ops.Index.IndexWorkspace(context.Background(), ops.Root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}

	summary, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: "runbooks/deploy.md"},
		Mode:        mcpschema.ReadModeSummary,
	})
	if err != nil {
		t.Fatalf("summary ReadDocument returned error: %v", err)
	}
	if summary.Data.Mode != mcpschema.ReadModeSummary || summary.Data.Summary == "" || len(summary.Data.TOC) == 0 {
		t.Fatalf("unexpected summary response: %#v", summary)
	}

	frontmatter, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Slug: "deploy-runbook"},
		Mode:        mcpschema.ReadModeFrontmatter,
	})
	if err != nil {
		t.Fatalf("frontmatter ReadDocument returned error: %v", err)
	}
	if frontmatter.Data.Frontmatter["title"] != "Deploy Runbook" {
		t.Fatalf("unexpected frontmatter: %#v", frontmatter.Data.Frontmatter)
	}

	toc, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{ID: "cairn:deploy"},
		Mode:        mcpschema.ReadModeTOC,
	})
	if err != nil {
		t.Fatalf("toc ReadDocument returned error: %v", err)
	}
	if len(toc.Data.TOC) != 3 || toc.Data.TOC[1].Text != "Steps" || toc.Data.TOC[1].ID != "steps" {
		t.Fatalf("unexpected toc: %#v", toc.Data.TOC)
	}

	sections, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: "runbooks/deploy.md"},
		Mode:        mcpschema.ReadModeSections,
		Sections:    []string{"Steps"},
	})
	if err != nil {
		t.Fatalf("sections ReadDocument returned error: %v", err)
	}
	if len(sections.Data.Sections) != 1 || sections.Data.Sections[0].Heading != "Steps" || sections.Data.Sections[0].Content == "" {
		t.Fatalf("unexpected sections: %#v", sections.Data.Sections)
	}

	full, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: "runbooks/deploy.md"},
		Mode:        mcpschema.ReadModeFull,
	})
	if err != nil {
		t.Fatalf("full ReadDocument returned error: %v", err)
	}
	if full.Data.Content == "" || len(full.NextSteps) == 0 {
		t.Fatalf("expected full content and next-step guidance: %#v", full)
	}
}

func TestReadDocumentWarnsForMissingSection(t *testing.T) {
	ops := newFixtureOps(t)
	envelope, err := ops.ReadDocument(context.Background(), mcpschema.ReadDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: "runbooks/auth.md"},
		Mode:        mcpschema.ReadModeSections,
		Sections:    []string{"Missing"},
	})
	if err != nil {
		t.Fatalf("ReadDocument returned error: %v", err)
	}
	if len(envelope.Data.Sections) != 0 || len(envelope.Warnings) != 1 {
		t.Fatalf("expected missing section warning: %#v", envelope)
	}
	if envelope.Warnings[0].Details["heading"] != "Missing" {
		t.Fatalf("unexpected warning details: %#v", envelope.Warnings[0])
	}
}
