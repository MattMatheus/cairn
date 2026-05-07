package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/syncstate"
)

func TestRunInitValidateAndSearch(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, code := run(t, "--root", root, "init", "--workspace-id", "cairn:workspace:test")
	if code != 0 {
		t.Fatalf("init code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Initialized workspace cairn:workspace:test") {
		t.Fatalf("unexpected init stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "capture", "--actor", "codex", "--title", "Searchable Note", "--body", "alpha beta body")
	if code != 0 {
		t.Fatalf("capture code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Captured agents/codex/searchable-note.md") {
		t.Fatalf("unexpected capture stdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Next: promote the document") {
		t.Fatalf("capture should include next steps:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "validate", "agents/codex/searchable-note.md")
	if code != 0 {
		t.Fatalf("validate code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Workspace validation passed.") {
		t.Fatalf("unexpected validate stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "search", "--query", "alpha")
	if code != 0 {
		t.Fatalf("search code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "agents/codex/searchable-note.md") {
		t.Fatalf("unexpected search stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, "remote indexer is unavailable") || strings.Contains(stdout, "semantic search is unavailable") {
		t.Fatalf("auto local search should not warn about unconfigured remote indexer:\n%s", stdout)
	}
}

func TestRunSetupLocalSyncCreatesConfig(t *testing.T) {
	root := t.TempDir()
	remoteRoot := filepath.Join(t.TempDir(), "remote")

	stdout, stderr, code := run(t, "--root", root, "setup", "local-sync", "--workspace-id", "cairn:workspace:test", "--remote-root", remoteRoot)
	if code != 0 {
		t.Fatalf("setup failed code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Configured local sync in .cairn/config.yaml") {
		t.Fatalf("unexpected setup stdout:\n%s", stdout)
	}
	config := readFile(t, root, ".cairn/config.yaml")
	if !strings.Contains(config, "remote_sync:") || !strings.Contains(config, "provider: local_fs") || !strings.Contains(config, "root: "+strconv.Quote(remoteRoot)) {
		t.Fatalf("config missing local sync block:\n%s", config)
	}
}

func TestRunSetupAzureSyncCreatesConfig(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, code := run(t, "--root", root, "setup", "azure-sync", "--workspace-id", "cairn:workspace:test", "--account", "cairnpilot", "--container", "pod-a")
	if code != 0 {
		t.Fatalf("setup failed code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Configured Azure Blob sync in .cairn/config.yaml") || !strings.Contains(stdout, "Prefix: <none>") {
		t.Fatalf("unexpected setup stdout:\n%s", stdout)
	}
	config := readFile(t, root, ".cairn/config.yaml")
	for _, expected := range []string{"remote_sync:", "provider: azure_blob", "account: \"cairnpilot\"", "container: \"pod-a\""} {
		if !strings.Contains(config, expected) {
			t.Fatalf("config missing %q:\n%s", expected, config)
		}
	}
}

func TestRunVersionAndDoctor(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, code := run(t, "version")
	if code != 0 {
		t.Fatalf("version code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "cairn dev") {
		t.Fatalf("unexpected version stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "doctor")
	if code != 0 {
		t.Fatalf("doctor code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Config: missing") {
		t.Fatalf("unexpected doctor stdout:\n%s", stdout)
	}

	runOK(t, "--root", root, "setup", "local-sync", "--remote-root", filepath.Join(t.TempDir(), "remote"))
	stdout, stderr, code = run(t, "--root", root, "doctor")
	if code != 0 {
		t.Fatalf("doctor configured code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Config: present") || !strings.Contains(stdout, "Remote sync: local_fs") {
		t.Fatalf("unexpected doctor configured stdout:\n%s", stdout)
	}
}

func TestRunPromoteArchiveAndIndexStatus(t *testing.T) {
	root := t.TempDir()
	runOK(t, "--root", root, "init")
	runOK(t, "--root", root, "capture", "--actor", "codex", "--title", "Promote Me", "--body", "body")

	stdout, stderr, code := run(t, "--root", root, "promote", "--path", "agents/codex/promote-me.md", "--type", "runbook")
	if code != 0 {
		t.Fatalf("promote code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Promoted runbooks/promote-me.md") {
		t.Fatalf("unexpected promote stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "archive", "runbooks/promote-me.md")
	if code != 0 {
		t.Fatalf("archive code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Archived archive/runbooks/promote-me.md") {
		t.Fatalf("unexpected archive stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "purge", "archive/runbooks/promote-me.md")
	if code == 0 {
		t.Fatalf("expected purge without confirmation to fail, stdout=%s", stdout)
	}
	if !strings.Contains(stderr, "requires --confirm-purge") {
		t.Fatalf("unexpected purge stderr:\n%s", stderr)
	}
	if _, err := os.Stat(filepath.Join(root, "archive", "runbooks", "promote-me.md")); err != nil {
		t.Fatalf("expected unconfirmed purge to keep file: %v", err)
	}

	stdout, stderr, code = run(t, "--root", root, "purge", "--confirm-purge", "archive/runbooks/promote-me.md")
	if code != 0 {
		t.Fatalf("purge code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Purged archive/runbooks/promote-me.md") || !strings.Contains(stdout, "Next:") {
		t.Fatalf("unexpected purge stdout:\n%s", stdout)
	}
	if _, err := os.Stat(filepath.Join(root, "archive", "runbooks", "promote-me.md")); !os.IsNotExist(err) {
		t.Fatalf("expected purged file removed, stat err: %v", err)
	}

	stdout, stderr, code = run(t, "--root", root, "index", "status")
	if code != 0 {
		t.Fatalf("index status code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Local index available: true") {
		t.Fatalf("unexpected index status stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "index", "refresh")
	if code != 0 {
		t.Fatalf("index refresh code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Local index refreshed: true") || !strings.Contains(stdout, "Remote index refreshed: false") {
		t.Fatalf("unexpected index refresh stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "sync", "status")
	if code != 0 {
		t.Fatalf("sync status code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Sync diverged: false") {
		t.Fatalf("unexpected sync status stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "sync", "dry-run")
	if code != 0 {
		t.Fatalf("sync dry-run code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Sync dry-run direction: push") {
		t.Fatalf("unexpected sync dry-run stdout:\n%s", stdout)
	}
}

func TestRunSyncPushWithoutRemoteSuggestsSetup(t *testing.T) {
	root := t.TempDir()
	runOK(t, "--root", root, "init")

	_, stderr, code := run(t, "--root", root, "sync", "push")
	if code == 0 {
		t.Fatalf("expected sync push without remote to fail")
	}
	if !strings.Contains(stderr, "remote sync is not configured") || !strings.Contains(stderr, "setup azure-sync") {
		t.Fatalf("unexpected stderr:\n%s", stderr)
	}
}

func TestRunSyncDryRunReportsDeleteAfterPurge(t *testing.T) {
	root := t.TempDir()
	runOK(t, "--root", root, "init")
	runOK(t, "--root", root, "capture", "--actor", "codex", "--title", "Delete Me", "--body", "body")
	runOK(t, "--root", root, "promote", "--path", "agents/codex/delete-me.md", "--type", "runbook")
	runOK(t, "--root", root, "archive", "runbooks/delete-me.md")

	saveSyncedManifestFixture(t, root)
	runOK(t, "--root", root, "purge", "--confirm-purge", "archive/runbooks/delete-me.md")

	stdout, stderr, code := run(t, "--root", root, "sync", "dry-run")
	if code != 0 {
		t.Fatalf("sync dry-run code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Sync dry-run direction: push") {
		t.Fatalf("unexpected sync dry-run stdout:\n%s", stdout)
	}
	if !strings.Contains(stdout, "- delete archive/runbooks/delete-me.md") {
		t.Fatalf("sync dry-run should report purge deletion:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Next: Push local changes when ready") {
		t.Fatalf("sync dry-run should report next step:\n%s", stdout)
	}
}

func TestRunCaptureReadsBodyFile(t *testing.T) {
	root := t.TempDir()
	bodyPath := filepath.Join(root, "body.txt")
	if err := os.WriteFile(bodyPath, []byte("from file"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stdout, stderr, code := run(t, "--root", root, "capture", "--actor", "codex", "--title", "File Body", "--body-file", bodyPath)
	if code != 0 {
		t.Fatalf("capture code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Captured agents/codex/file-body.md") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
	content, err := os.ReadFile(filepath.Join(root, "agents", "codex", "file-body.md"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "from file") {
		t.Fatalf("captured file missing body:\n%s", string(content))
	}
}

func saveSyncedManifestFixture(t *testing.T, root string) {
	t.Helper()
	now := func() time.Time { return time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC) }
	manifest, err := syncstate.Generate(root, syncstate.GenerateOptions{WorkspaceID: "pod-1", Now: now})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	hash, err := syncstate.Hash(manifest)
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if err := syncstate.Save(root, syncstate.State{
		LastRemoteManifestHash: hash,
		LastSyncAt:             now(),
		Entries:                manifest.Entries,
	}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	path := filepath.Join(root, ".cairn", "remote-manifest.json")
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	_, stderr, code := run(t, "nope")
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("unexpected stderr %q", stderr)
	}
}

func TestRunMCPReadonlyServesJSONRPC(t *testing.T) {
	root := t.TempDir()
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run(context.Background(), []string{"--root", root, "mcp", "readonly"}, strings.NewReader(input), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("mcp readonly code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "read_document") {
		t.Fatalf("unexpected mcp output:\n%s", stdout.String())
	}
}

func runOK(t *testing.T, args ...string) string {
	t.Helper()
	stdout, stderr, code := run(t, args...)
	if code != 0 {
		t.Fatalf("%v code=%d stderr=%s", args, code, stderr)
	}
	return stdout
}

func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func readFile(t *testing.T, root string, rel string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", rel, err)
	}
	return string(content)
}
