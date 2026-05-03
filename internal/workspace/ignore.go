package workspace

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ignoreFile struct {
	rules []ignoreRule
}

type ignoreRule struct {
	pattern string
	negate  bool
	dirOnly bool
	rooted  bool
}

func loadIgnore(root string) (ignoreFile, error) {
	content, err := os.ReadFile(filepath.Join(root, ".cairnignore"))
	if errors.Is(err, os.ErrNotExist) {
		return ignoreFile{}, nil
	}
	if err != nil {
		return ignoreFile{}, err
	}

	var ignore ignoreFile
	for _, line := range strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		rule := ignoreRule{}
		if strings.HasPrefix(trimmed, "!") {
			rule.negate = true
			trimmed = strings.TrimPrefix(trimmed, "!")
		}
		if strings.HasPrefix(trimmed, "/") {
			rule.rooted = true
			trimmed = strings.TrimPrefix(trimmed, "/")
		}
		if strings.HasSuffix(trimmed, "/") {
			rule.dirOnly = true
			trimmed = strings.TrimSuffix(trimmed, "/")
		}
		if trimmed == "" {
			continue
		}
		rule.pattern = filepath.ToSlash(trimmed)
		ignore.rules = append(ignore.rules, rule)
	}
	return ignore, nil
}

func (ignore ignoreFile) matches(relativePath string, isDir bool) bool {
	relativePath = path.Clean(filepath.ToSlash(relativePath))
	matched := false
	for _, rule := range ignore.rules {
		if rule.matches(relativePath, isDir) {
			matched = !rule.negate
		}
	}
	return matched
}

func (rule ignoreRule) matches(relativePath string, isDir bool) bool {
	if rule.dirOnly && !isDir && !strings.HasPrefix(relativePath, rule.pattern+"/") {
		return false
	}
	if rule.rooted || strings.Contains(rule.pattern, "/") {
		return pathMatch(rule.pattern, relativePath) || strings.HasPrefix(relativePath, rule.pattern+"/")
	}
	if pathMatch(rule.pattern, path.Base(relativePath)) {
		return true
	}
	for _, part := range strings.Split(relativePath, "/") {
		if pathMatch(rule.pattern, part) {
			return true
		}
	}
	return false
}

func pathMatch(pattern string, value string) bool {
	ok, err := path.Match(pattern, value)
	if err != nil {
		return pattern == value
	}
	if ok {
		return true
	}
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			return strings.HasPrefix(value, parts[0]) && strings.HasSuffix(value, parts[1])
		}
	}
	return pattern == value
}
