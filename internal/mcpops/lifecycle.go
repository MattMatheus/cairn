package mcpops

import (
	"context"
	"path/filepath"

	"cairn/internal/document"
	"cairn/internal/mcpschema"
)

func (l *Local) CaptureNote(_ context.Context, req mcpschema.CaptureNoteRequest) (mcpschema.Envelope[mcpschema.MutationResult], error) {
	workspace := l.documentWorkspace()
	result, err := workspace.Capture(document.CaptureOptions{
		Actor:   req.Actor,
		Title:   req.Title,
		Body:    req.Body,
		Type:    req.Type,
		Authors: req.Authors,
		Tags:    req.Tags,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.MutationResult]{}, err
	}
	return l.mutationEnvelope(result, "created"), nil
}

func (l *Local) PromoteDocument(_ context.Context, req mcpschema.PromoteDocumentRequest) (mcpschema.Envelope[mcpschema.MutationResult], error) {
	workspace := l.documentWorkspace()
	result, err := workspace.Promote(document.PromoteOptions{
		Path:   req.Path,
		Type:   req.Type,
		Status: req.Status,
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.MutationResult]{}, err
	}
	return l.mutationEnvelope(result, "promoted"), nil
}

func (l *Local) ArchiveDocument(_ context.Context, req mcpschema.ArchiveDocumentRequest) (mcpschema.Envelope[mcpschema.MutationResult], error) {
	workspace := l.documentWorkspace()
	result, err := workspace.Archive(document.ArchiveOptions{Path: req.Path})
	if err != nil {
		return mcpschema.Envelope[mcpschema.MutationResult]{}, err
	}
	return l.mutationEnvelope(result, "archived"), nil
}

func (l *Local) documentWorkspace() document.Workspace {
	return document.Workspace{
		Root: l.Root,
		Now:  l.Now,
	}
}

func (l *Local) mutationEnvelope(result document.OperationResult, kind string) mcpschema.Envelope[mcpschema.MutationResult] {
	nextSteps := make([]mcpschema.NextStep, 0, len(result.NextSteps))
	for _, step := range result.NextSteps {
		nextSteps = append(nextSteps, mcpschema.NextStep{
			Action: "next",
			Label:  step,
		})
	}
	changed := mcpschema.ChangedPath{
		Path: filepath.ToSlash(result.Path),
		Kind: kind,
	}
	if result.OriginalPath != "" {
		changed.PreviousPath = filepath.ToSlash(result.OriginalPath)
	}
	return mcpschema.Envelope[mcpschema.MutationResult]{
		OK: true,
		Data: mcpschema.MutationResult{
			DocumentID:   result.DocumentID,
			ChangedPaths: []mcpschema.ChangedPath{changed},
		},
		NextSteps:  nextSteps,
		Provenance: l.provenance("document_lifecycle"),
	}
}
