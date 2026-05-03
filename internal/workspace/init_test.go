package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesStandardWorkspaceLayout(t *testing.T) {
	root := t.TempDir()

	result, err := Init(root, InitOptions{WorkspaceID: "cairn:workspace:test"})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if result.WorkspaceID != "cairn:workspace:test" {
		t.Fatalf("unexpected workspace id %q", result.WorkspaceID)
	}

	for _, dir := range standardFolders {
		assertDir(t, root, dir)
	}
	for _, file := range []string{
		".cairn/config.yaml",
		".cairnignore",
		".cairn/schemas/core.yaml",
		".cairn/schemas/README.md",
		"onboarding/team-context.md",
		"onboarding/agent-setup.md",
		"onboarding/workspace-map.md",
		"AGENTS.md",
		"CLAUDE.md",
	} {
		assertFile(t, root, file)
	}

	config := readFile(t, root, ".cairn/config.yaml")
	for _, expected := range []string{
		"schema_version: 1",
		"workspace_id: cairn:workspace:test",
		"managed_folders:",
		"document_types:",
		"pod-remote:",
	} {
		if !strings.Contains(config, expected) {
			t.Fatalf("config missing %q:\n%s", expected, config)
		}
	}
	if len(result.Created) == 0 {
		t.Fatalf("expected created paths")
	}
	for _, file := range []string{
		"onboarding/team-context.md",
		"onboarding/agent-setup.md",
		"onboarding/workspace-map.md",
	} {
		if !strings.Contains(readFile(t, root, file), "type: onboarding") {
			t.Fatalf("%s missing starter onboarding frontmatter", file)
		}
	}
}

func TestInitIsIdempotentAndNonDestructive(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "AGENTS.md", "# Custom Agents\n")
	writeFile(t, root, ".cairn/config.yaml", "custom: true\n")

	first, err := Init(root, InitOptions{WorkspaceID: "cairn:workspace:first"})
	if err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	second, err := Init(root, InitOptions{WorkspaceID: "cairn:workspace:second"})
	if err != nil {
		t.Fatalf("second Init() error = %v", err)
	}

	if readFile(t, root, "AGENTS.md") != "# Custom Agents\n" {
		t.Fatalf("AGENTS.md was overwritten")
	}
	if readFile(t, root, ".cairn/config.yaml") != "custom: true\n" {
		t.Fatalf("config was overwritten")
	}
	if len(second.Created) != 0 {
		t.Fatalf("expected second init to create no paths, got %#v", second.Created)
	}
	if len(second.Existing) <= len(first.Existing) {
		t.Fatalf("expected second init to report existing paths")
	}
}

func TestInitCreatesWorkspaceIDWhenMissing(t *testing.T) {
	root := t.TempDir()

	result, err := Init(root, InitOptions{})
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if !strings.HasPrefix(result.WorkspaceID, "cairn:workspace:") {
		t.Fatalf("unexpected generated workspace id %q", result.WorkspaceID)
	}
	if !strings.Contains(readFile(t, root, ".cairn/config.yaml"), result.WorkspaceID) {
		t.Fatalf("config does not contain generated workspace id")
	}
}

func TestInitRefusesConflictingPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "inbox", "not a directory\n")

	if _, err := Init(root, InitOptions{}); err == nil {
		t.Fatalf("expected conflicting path error")
	}
}

func assertDir(t *testing.T, root string, rel string) {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("Stat(%s) error = %v", rel, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", rel)
	}
}

func assertFile(t *testing.T, root string, rel string) {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("Stat(%s) error = %v", rel, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory", rel)
	}
}

func readFile(t *testing.T, root string, rel string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", rel, err)
	}
	return string(content)
}
