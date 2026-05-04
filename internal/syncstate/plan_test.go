package syncstate

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanFromStatusClean(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	plan := PlanFromStatus(status)
	if plan.Direction != PlanDirectionClean || !plan.Safe || len(plan.Changes) != 0 {
		t.Fatalf("unexpected clean plan %#v", plan)
	}
}

func TestPlanFromStatusLocalOnlyPush(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	writeFile(t, root, "working/local.md", managedDocument("cairn:local", "working", "note", "Local"))

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	plan := PlanFromStatus(status)
	if plan.Direction != PlanDirectionPush || !plan.Safe {
		t.Fatalf("unexpected push plan %#v", plan)
	}
	assertChange(t, plan.Changes, ChangeCreate, "working/local.md")
}

func TestPlanFromStatusRemoteOnlyPull(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	remote := base
	remote.Entries = append(remote.Entries, Entry{Path: "working/remote.md", Kind: "file", Size: 1, Hash: "remote", DocumentID: "cairn:remote", Status: "working", Type: "note"})
	writeRemoteManifest(t, root, remote)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	plan := PlanFromStatus(status)
	if plan.Direction != PlanDirectionPull || !plan.Safe {
		t.Fatalf("unexpected pull plan %#v", plan)
	}
	assertChange(t, plan.Changes, ChangeCreate, "working/remote.md")
}

func TestPlanFromStatusDivergedRefusedDoesNotMutate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A"))
	base := mustGenerate(t, root)
	saveBaseAndRemote(t, root, base)
	beforeState := readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))

	writeFile(t, root, "working/a.md", managedDocument("cairn:a", "working", "note", "A Local"))
	localContent := readBytes(t, filepath.Join(root, "working", "a.md"))
	remote := base
	for i := range remote.Entries {
		if remote.Entries[i].Path == "working/a.md" {
			remote.Entries[i].Hash = "remote"
			remote.Entries[i].Size++
		}
	}
	writeRemoteManifest(t, root, remote)

	status, err := StatusReport(context.Background(), root, StatusOptions{WorkspaceID: "pod-1"})
	if err != nil {
		t.Fatalf("StatusReport() error = %v", err)
	}
	envelope := PlanEnvelope(status)
	if envelope.OK || envelope.Data.Plan == nil || envelope.Data.Plan.Direction != "refused" {
		t.Fatalf("unexpected refused envelope %#v", envelope)
	}
	if len(envelope.Data.Plan.Conflicts) == 0 || len(envelope.Warnings) == 0 {
		t.Fatalf("expected conflicts and warnings: %#v", envelope)
	}
	if string(beforeState) != string(readBytes(t, filepath.Join(root, ".cairn", "sync-state.json"))) {
		t.Fatalf("dry-run mutated sync state")
	}
	if string(localContent) != string(readBytes(t, filepath.Join(root, "working", "a.md"))) || !strings.Contains(string(localContent), "A Local") {
		t.Fatalf("dry-run mutated local file")
	}
}
