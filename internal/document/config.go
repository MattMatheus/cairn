package document

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	SchemaVersion  int
	WorkspaceID    string
	ManagedFolders []string
	DocumentTypes  map[string]string
}

func DefaultConfig() Config {
	return Config{
		SchemaVersion: 1,
		ManagedFolders: []string{
			"inbox",
			"agents",
			"working",
			"decisions",
			"runbooks",
			"projects",
			"services",
			"handoffs",
			"onboarding",
			"archive",
		},
		DocumentTypes: map[string]string{
			"note":          "working",
			"handoff":       "handoffs",
			"investigation": "working",
			"decision":      "decisions",
			"runbook":       "runbooks",
			"project":       "projects",
			"service":       "services",
			"onboarding":    "onboarding",
		},
	}
}

func LoadConfig(root string) (Config, error) {
	cfg := DefaultConfig()
	content, err := os.ReadFile(filepath.Join(root, ".cairn", "config.yaml"))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}
	return parseConfig(string(content), cfg), nil
}

func (c Config) DestinationFolder(docType string) string {
	if folder := c.DocumentTypes[docType]; folder != "" {
		return cleanConfigPath(folder)
	}
	if folder := DefaultConfig().DocumentTypes[docType]; folder != "" {
		return folder
	}
	return "working"
}

func (c Config) ManagedFolderSet() map[string]struct{} {
	out := map[string]struct{}{}
	for _, folder := range c.ManagedFolders {
		if cleaned := cleanConfigPath(folder); cleaned != "" {
			out[cleaned] = struct{}{}
		}
	}
	return out
}

func parseConfig(content string, cfg Config) Config {
	section := ""
	for _, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(raw, " ") && strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		if !strings.HasPrefix(raw, " ") {
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			switch strings.TrimSpace(key) {
			case "workspace_id":
				cfg.WorkspaceID = unquoteConfig(strings.TrimSpace(value))
			case "schema_version":
				if strings.TrimSpace(value) == "1" {
					cfg.SchemaVersion = 1
				}
			}
			section = ""
			continue
		}
		switch section {
		case "managed_folders":
			if strings.HasPrefix(line, "- ") {
				folder := cleanConfigPath(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
				if folder != "" {
					cfg.ManagedFolders = appendUniqueString(cfg.ManagedFolders, folder)
				}
			}
		case "document_types":
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			docType := strings.TrimSpace(key)
			folder := cleanConfigPath(unquoteConfig(strings.TrimSpace(value)))
			if docType != "" && folder != "" {
				if cfg.DocumentTypes == nil {
					cfg.DocumentTypes = map[string]string{}
				}
				cfg.DocumentTypes[docType] = folder
			}
		}
	}
	sort.Strings(cfg.ManagedFolders)
	return cfg
}

func appendUniqueString(values []string, next string) []string {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}

func cleanConfigPath(value string) string {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimPrefix(strings.TrimSpace(value), "/")))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ""
	}
	return clean
}

func unquoteConfig(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)
	value = strings.Trim(value, `'`)
	return value
}
