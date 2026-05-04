package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(stdout, "Workspace validation passed with warnings.") {
		t.Fatalf("unexpected validate stdout:\n%s", stdout)
	}

	stdout, stderr, code = run(t, "--root", root, "search", "--query", "alpha")
	if code != 0 {
		t.Fatalf("search code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "agents/codex/searchable-note.md") {
		t.Fatalf("unexpected search stdout:\n%s", stdout)
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

	stdout, stderr, code = run(t, "--root", root, "index", "status")
	if code != 0 {
		t.Fatalf("index status code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Local index available: true") {
		t.Fatalf("unexpected index status stdout:\n%s", stdout)
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
