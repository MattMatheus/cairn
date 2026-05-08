package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachRepoRecordsMetadataAndPointer(t *testing.T) {
	podRoot := t.TempDir()
	repoRoot := filepath.Join(filepath.Dir(podRoot), "payments-api")
	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if _, err := Init(podRoot, InitOptions{WorkspaceID: "cairn:workspace:test"}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	result, err := AttachRepo(podRoot, RepoAttachOptions{
		Name:         "payments-api",
		Path:         "../payments-api",
		URL:          "https://dev.azure.com/org/project/_git/payments-api",
		WritePointer: true,
	})
	if err != nil {
		t.Fatalf("AttachRepo() error = %v", err)
	}
	if result.Repo.Name != "payments-api" || result.Repo.Path != "../payments-api" {
		t.Fatalf("unexpected repo result: %#v", result.Repo)
	}
	content := readWorkspaceFile(t, podRoot, ".cairn/repos.yaml")
	for _, expected := range []string{
		"schema_version: 1",
		`name: "payments-api"`,
		`path: "../payments-api"`,
		`url: "https://dev.azure.com/org/project/_git/payments-api"`,
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("repos.yaml missing %q:\n%s", expected, content)
		}
	}
	pointer := readAbsoluteFile(t, filepath.Join(repoRoot, RepoPointerFile))
	if strings.TrimSpace(pointer) != "../"+filepath.Base(podRoot) {
		t.Fatalf("unexpected pointer content %q", pointer)
	}

	discovered, err := DiscoverWorkspace(repoRoot)
	if err != nil {
		t.Fatalf("DiscoverWorkspace() error = %v", err)
	}
	if discovered.WorkspacePath != podRoot {
		t.Fatalf("unexpected workspace path: %s", discovered.WorkspacePath)
	}
}

func TestLoadReposSortsByName(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cairn"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".cairn", "repos.yaml"), []byte(`schema_version: 1
repos:
  - name: "worker"
    path: "../worker"
  - name: "api"
    path: "../api"
`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	repos, err := LoadRepos(root)
	if err != nil {
		t.Fatalf("LoadRepos() error = %v", err)
	}
	if len(repos) != 2 || repos[0].Name != "api" || repos[1].Name != "worker" {
		t.Fatalf("repos not sorted: %#v", repos)
	}
}

func TestDiscoverWorkspaceRequiresPointer(t *testing.T) {
	_, err := DiscoverWorkspace(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "no .cairn-workspace pointer") {
		t.Fatalf("expected missing pointer error, got %v", err)
	}
}

func readWorkspaceFile(t *testing.T, root string, rel string) string {
	t.Helper()
	return readAbsoluteFile(t, filepath.Join(root, filepath.FromSlash(rel)))
}

func readAbsoluteFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}
