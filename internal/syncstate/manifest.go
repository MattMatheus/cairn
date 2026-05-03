package syncstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"cairn/internal/document"
)

const ManifestVersion = 1

type Manifest struct {
	ManifestVersion int       `json:"manifest_version"`
	GeneratedAt     time.Time `json:"generated_at"`
	WorkspaceID     string    `json:"workspace_id,omitempty"`
	Entries         []Entry   `json:"entries"`
}

type Entry struct {
	Path       string    `json:"path"`
	Kind       string    `json:"kind"`
	Size       int64     `json:"size"`
	Hash       string    `json:"hash"`
	Modified   time.Time `json:"modified"`
	DocumentID string    `json:"document_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	Type       string    `json:"type,omitempty"`
}

type GenerateOptions struct {
	WorkspaceID string
	Now         func() time.Time
}

func Generate(root string, opts GenerateOptions) (Manifest, error) {
	ignore, err := loadIgnore(root)
	if err != nil {
		return Manifest{}, err
	}

	generatedAt := time.Now().UTC()
	if opts.Now != nil {
		generatedAt = opts.Now().UTC()
	}
	manifest := Manifest{
		ManifestVersion: ManifestVersion,
		GeneratedAt:     generatedAt,
		WorkspaceID:     opts.WorkspaceID,
	}

	err = filepath.WalkDir(root, func(absolutePath string, dirEntry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if absolutePath == root {
			return nil
		}

		rel, err := filepath.Rel(root, absolutePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".git" || strings.HasPrefix(rel, ".git/") {
			if dirEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if ignore.matches(rel, dirEntry.IsDir()) {
			if dirEntry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if dirEntry.IsDir() {
			return nil
		}

		entry, err := manifestEntry(absolutePath, rel)
		if err != nil {
			return err
		}
		manifest.Entries = append(manifest.Entries, entry)
		return nil
	})
	if err != nil {
		return Manifest{}, err
	}

	normalizeEntries(manifest.Entries)
	return manifest, nil
}

func Hash(manifest Manifest) (string, error) {
	normalized := manifest
	normalized.GeneratedAt = time.Time{}
	normalizeEntries(normalized.Entries)
	content, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func manifestEntry(absolutePath string, relativePath string) (Entry, error) {
	info, err := os.Stat(absolutePath)
	if err != nil {
		return Entry{}, err
	}
	file, err := os.Open(absolutePath)
	if err != nil {
		return Entry{}, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return Entry{}, err
	}
	entry := Entry{
		Path:     relativePath,
		Kind:     "file",
		Size:     info.Size(),
		Hash:     hex.EncodeToString(hasher.Sum(nil)),
		Modified: info.ModTime().UTC().Truncate(time.Second),
	}

	if strings.EqualFold(filepath.Ext(relativePath), ".md") {
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			return Entry{}, err
		}
		parsed, err := document.ParseMarkdown(string(content))
		if err == nil && parsed.HasFrontmatter {
			entry.DocumentID = parsed.Metadata.ID
			entry.Status = parsed.Metadata.Status
			entry.Type = parsed.Metadata.Type
		}
	}

	return entry, nil
}

func normalizeEntries(entries []Entry) {
	for i := range entries {
		entries[i].Path = path.Clean(strings.TrimPrefix(filepath.ToSlash(entries[i].Path), "/"))
		if entries[i].Path == "." {
			entries[i].Path = ""
		}
		entries[i].Modified = entries[i].Modified.UTC().Truncate(time.Second)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
}

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
