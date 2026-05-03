package mcpops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/mcpschema"
)

func TestLifecycleMutationAdaptersCapturePromoteAndArchive(t *testing.T) {
	ops := newLifecycleOps(t)

	captured, err := ops.CaptureNote(context.Background(), mcpschema.CaptureNoteRequest{
		ActorContext: mcpschema.ActorContext{Actor: "codex"},
		Title:        "Adapter Capture",
		Body:         "Captured through MCP adapter.",
		Type:         "investigation",
		Tags:         []string{"mcp"},
	})
	if err != nil {
		t.Fatalf("CaptureNote returned error: %v", err)
	}
	if !captured.OK || captured.Data.DocumentID == "" || len(captured.Data.ChangedPaths) != 1 || len(captured.NextSteps) == 0 {
		t.Fatalf("unexpected capture envelope: %#v", captured)
	}
	capturedPath := captured.Data.ChangedPaths[0].Path
	if capturedPath != "agents/codex/adapter-capture.md" {
		t.Fatalf("unexpected capture path: %s", capturedPath)
	}
	if _, err := os.Stat(filepath.Join(ops.Root, filepath.FromSlash(capturedPath))); err != nil {
		t.Fatalf("expected captured file: %v", err)
	}

	promoted, err := ops.PromoteDocument(context.Background(), mcpschema.PromoteDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: capturedPath},
		Type:        "runbook",
		Status:      "proposed",
	})
	if err != nil {
		t.Fatalf("PromoteDocument returned error: %v", err)
	}
	if promoted.Data.DocumentID != captured.Data.DocumentID ||
		promoted.Data.ChangedPaths[0].PreviousPath != capturedPath ||
		promoted.Data.ChangedPaths[0].Kind != "promoted" {
		t.Fatalf("unexpected promote envelope: %#v", promoted)
	}

	archived, err := ops.ArchiveDocument(context.Background(), mcpschema.ArchiveDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: promoted.Data.ChangedPaths[0].Path},
		Reason:      "done",
	})
	if err != nil {
		t.Fatalf("ArchiveDocument returned error: %v", err)
	}
	if archived.Data.DocumentID != captured.Data.DocumentID ||
		archived.Data.ChangedPaths[0].Kind != "archived" ||
		archived.Data.ChangedPaths[0].Path != "archive/runbooks/adapter-capture.md" {
		t.Fatalf("unexpected archive envelope: %#v", archived)
	}
	if len(archived.NextSteps) == 0 {
		t.Fatalf("expected archive next steps: %#v", archived)
	}
}

func TestLifecycleMutationAdaptersPreserveValidationFailures(t *testing.T) {
	ops := newLifecycleOps(t)
	if _, err := ops.CaptureNote(context.Background(), mcpschema.CaptureNoteRequest{
		ActorContext: mcpschema.ActorContext{Actor: "../bad"},
		Title:        "Unsafe",
		Body:         "Nope",
	}); err == nil {
		t.Fatal("expected unsafe actor validation error")
	}
	if _, err := ops.PromoteDocument(context.Background(), mcpschema.PromoteDocumentRequest{
		DocumentRef: mcpschema.DocumentRef{Path: "missing.md"},
		Status:      "canonical",
	}); err == nil {
		t.Fatal("expected missing path error")
	}
}

func newLifecycleOps(t *testing.T) *Local {
	t.Helper()
	root := t.TempDir()
	index, err := OpenLocal(root)
	if err != nil {
		t.Fatalf("OpenLocal returned error: %v", err)
	}
	t.Cleanup(func() { index.Close() })
	index.Now = func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) }
	return index
}
