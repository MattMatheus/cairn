package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const RepoPointerFile = ".cairn-workspace"

type RepoRef struct {
	Name string
	Path string
	URL  string
}

type RepoAttachOptions struct {
	Name         string
	Path         string
	URL          string
	WritePointer bool
}

type RepoAttachResult struct {
	Repo        RepoRef
	ConfigPath  string
	PointerPath string
}

type RepoDiscoverResult struct {
	WorkspacePath string
	PointerPath   string
}

func AttachRepo(root string, opts RepoAttachOptions) (RepoAttachResult, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return RepoAttachResult{}, errors.New("repo name is required")
	}
	if err := validateRepoName(name); err != nil {
		return RepoAttachResult{}, err
	}
	repoPath := strings.TrimSpace(opts.Path)
	if repoPath == "" {
		return RepoAttachResult{}, errors.New("repo path is required")
	}
	if filepath.IsAbs(repoPath) {
		return RepoAttachResult{}, fmt.Errorf("repo path must be relative to the Cairn workspace: %s", repoPath)
	}
	cleanRepoPath := cleanRepoPath(repoPath)
	if cleanRepoPath == "" {
		return RepoAttachResult{}, fmt.Errorf("repo path is invalid: %s", repoPath)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return RepoAttachResult{}, err
	}
	if _, err := os.Stat(filepath.Join(absRoot, ".cairn", "config.yaml")); err != nil {
		if os.IsNotExist(err) {
			return RepoAttachResult{}, errors.New("Cairn workspace config is missing; run `cairn init` first")
		}
		return RepoAttachResult{}, err
	}
	absRepo := filepath.Clean(filepath.Join(absRoot, filepath.FromSlash(cleanRepoPath)))
	if info, err := os.Stat(absRepo); err != nil {
		return RepoAttachResult{}, err
	} else if !info.IsDir() {
		return RepoAttachResult{}, fmt.Errorf("repo path is not a directory: %s", cleanRepoPath)
	}

	repos, err := LoadRepos(absRoot)
	if err != nil {
		return RepoAttachResult{}, err
	}
	next := RepoRef{Name: name, Path: cleanRepoPath, URL: strings.TrimSpace(opts.URL)}
	replaced := false
	for index, repo := range repos {
		if repo.Name == name {
			repos[index] = next
			replaced = true
			break
		}
	}
	if !replaced {
		repos = append(repos, next)
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	configPath := filepath.Join(absRoot, ".cairn", "repos.yaml")
	if err := SaveRepos(absRoot, repos); err != nil {
		return RepoAttachResult{}, err
	}

	result := RepoAttachResult{Repo: next, ConfigPath: configPath}
	if opts.WritePointer {
		relFromRepo, err := filepath.Rel(absRepo, absRoot)
		if err != nil {
			return RepoAttachResult{}, err
		}
		pointerPath := filepath.Join(absRepo, RepoPointerFile)
		if err := os.WriteFile(pointerPath, []byte(filepath.ToSlash(relFromRepo)+"\n"), 0o644); err != nil {
			return RepoAttachResult{}, err
		}
		result.PointerPath = pointerPath
	}
	return result, nil
}

func LoadRepos(root string) ([]RepoRef, error) {
	content, err := os.ReadFile(filepath.Join(root, ".cairn", "repos.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return parseRepos(string(content)), nil
}

func SaveRepos(root string, repos []RepoRef) error {
	if err := os.MkdirAll(filepath.Join(root, ".cairn"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, ".cairn", "repos.yaml"), []byte(renderRepos(repos)), 0o644)
}

func DiscoverWorkspace(from string) (RepoDiscoverResult, error) {
	start, err := filepath.Abs(from)
	if err != nil {
		return RepoDiscoverResult{}, err
	}
	info, err := os.Stat(start)
	if err != nil {
		return RepoDiscoverResult{}, err
	}
	if !info.IsDir() {
		start = filepath.Dir(start)
	}
	current := start
	for {
		pointerPath := filepath.Join(current, RepoPointerFile)
		if content, err := os.ReadFile(pointerPath); err == nil {
			target := strings.TrimSpace(string(content))
			if target == "" {
				return RepoDiscoverResult{}, fmt.Errorf("%s is empty", pointerPath)
			}
			var workspacePath string
			if filepath.IsAbs(target) {
				workspacePath = filepath.Clean(target)
			} else {
				workspacePath = filepath.Clean(filepath.Join(current, filepath.FromSlash(target)))
			}
			if _, err := os.Stat(filepath.Join(workspacePath, ".cairn", "config.yaml")); err != nil {
				if os.IsNotExist(err) {
					return RepoDiscoverResult{}, fmt.Errorf("%s points to %s, but no Cairn workspace config was found", pointerPath, workspacePath)
				}
				return RepoDiscoverResult{}, err
			}
			return RepoDiscoverResult{WorkspacePath: workspacePath, PointerPath: pointerPath}, nil
		} else if err != nil && !os.IsNotExist(err) {
			return RepoDiscoverResult{}, err
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return RepoDiscoverResult{}, fmt.Errorf("no %s pointer found from %s", RepoPointerFile, start)
}

func parseRepos(content string) []RepoRef {
	var repos []RepoRef
	var current *RepoRef
	flush := func() {
		if current != nil && current.Name != "" {
			repos = append(repos, *current)
		}
		current = nil
	}
	for _, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || line == "schema_version: 1" || line == "repos:" {
			continue
		}
		if strings.HasPrefix(line, "- name:") {
			flush()
			current = &RepoRef{Name: unquoteRepoValue(strings.TrimSpace(strings.TrimPrefix(line, "- name:")))}
			continue
		}
		if current == nil {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = unquoteRepoValue(strings.TrimSpace(value))
		switch strings.TrimSpace(key) {
		case "path":
			current.Path = cleanRepoPath(value)
		case "url":
			current.URL = value
		}
	}
	flush()
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	return repos
}

func renderRepos(repos []RepoRef) string {
	var builder strings.Builder
	builder.WriteString("schema_version: 1\nrepos:\n")
	for _, repo := range repos {
		builder.WriteString("  - name: ")
		builder.WriteString(quoteRepoValue(repo.Name))
		builder.WriteString("\n    path: ")
		builder.WriteString(quoteRepoValue(repo.Path))
		builder.WriteString("\n")
		if repo.URL != "" {
			builder.WriteString("    url: ")
			builder.WriteString(quoteRepoValue(repo.URL))
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func cleanRepoPath(value string) string {
	value = filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
	if value == "." || value == "" {
		return ""
	}
	return value
}

func validateRepoName(value string) error {
	if strings.ContainsAny(value, `/\`) || value == "." || value == ".." {
		return fmt.Errorf("repo name must be a simple name: %s", value)
	}
	return nil
}

func quoteRepoValue(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func unquoteRepoValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}
