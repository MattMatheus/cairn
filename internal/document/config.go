package document

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	SchemaVersion  int
	WorkspaceID    string
	ManagedFolders []string
	DocumentTypes  map[string]string
	RemoteSync     RemoteSyncConfig
	RemoteIndex    RemoteIndexConfig
}

type RemoteSyncConfig struct {
	Provider  string
	Account   string
	Endpoint  string
	Container string
	Prefix    string
	Root      string
	AuthMode  string
}

type RemoteIndexConfig struct {
	URL      string
	Audience string
	TenantID string
}

type ConfigFinding struct {
	Path     string
	Severity string
	Message  string
}

var coreSchemaRequiredFields = []string{"id", "schema_version", "title", "slug", "type", "status", "created", "updated", "authors", "actors", "source", "tags"}

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

func ValidateConfigFiles(root string) []ConfigFinding {
	var findings []ConfigFinding
	findings = append(findings, validateConfigFile(root)...)
	findings = append(findings, validateSchemaFiles(root)...)
	return findings
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
	profile := ""
	podRemoteEnabled := false
	var podRemoteSync RemoteSyncConfig
	var podRemoteIndex RemoteIndexConfig
	for _, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(raw, " ") && strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			profile = ""
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
			profile = ""
			continue
		}
		if section == "profiles" {
			indent := leadingSpaces(raw)
			if indent == 2 && strings.HasSuffix(line, ":") {
				profile = strings.TrimSuffix(line, ":")
				continue
			}
			if indent >= 4 && profile == "pod-remote" {
				key, value, ok := strings.Cut(line, ":")
				if !ok {
					continue
				}
				value = unquoteConfig(strings.TrimSpace(value))
				switch strings.TrimSpace(key) {
				case "enabled":
					podRemoteEnabled = value == "true"
				case "provider":
					podRemoteSync.Provider = value
				case "account":
					podRemoteSync.Account = value
				case "endpoint":
					podRemoteSync.Endpoint = value
				case "container":
					podRemoteSync.Container = value
				case "prefix":
					podRemoteSync.Prefix = cleanConfigPath(value)
				case "root":
					podRemoteSync.Root = value
				case "auth_mode":
					podRemoteSync.AuthMode = value
				case "url":
					podRemoteIndex.URL = value
				case "audience":
					podRemoteIndex.Audience = value
				case "tenant_id":
					podRemoteIndex.TenantID = value
				}
			}
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
		case "remote_sync":
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			value = unquoteConfig(strings.TrimSpace(value))
			switch strings.TrimSpace(key) {
			case "provider":
				cfg.RemoteSync.Provider = value
			case "account":
				cfg.RemoteSync.Account = value
			case "endpoint":
				cfg.RemoteSync.Endpoint = value
			case "container":
				cfg.RemoteSync.Container = value
			case "prefix":
				cfg.RemoteSync.Prefix = cleanConfigPath(value)
			case "root":
				cfg.RemoteSync.Root = value
			case "auth_mode":
				cfg.RemoteSync.AuthMode = value
			}
		case "remote_index":
			key, value, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			value = unquoteConfig(strings.TrimSpace(value))
			switch strings.TrimSpace(key) {
			case "url":
				cfg.RemoteIndex.URL = value
			case "audience":
				cfg.RemoteIndex.Audience = value
			case "tenant_id":
				cfg.RemoteIndex.TenantID = value
			}
		}
	}
	if podRemoteEnabled {
		mergeRemoteSync(&cfg.RemoteSync, podRemoteSync)
		mergeRemoteIndex(&cfg.RemoteIndex, podRemoteIndex)
	}
	sort.Strings(cfg.ManagedFolders)
	return cfg
}

func mergeRemoteSync(target *RemoteSyncConfig, profile RemoteSyncConfig) {
	if profile.Provider != "" {
		target.Provider = profile.Provider
	}
	if profile.Account != "" {
		target.Account = profile.Account
	}
	if profile.Endpoint != "" {
		target.Endpoint = profile.Endpoint
	}
	if profile.Container != "" {
		target.Container = profile.Container
	}
	if profile.Prefix != "" {
		target.Prefix = profile.Prefix
	}
	if profile.Root != "" {
		target.Root = profile.Root
	}
	if profile.AuthMode != "" {
		target.AuthMode = profile.AuthMode
	}
}

func mergeRemoteIndex(target *RemoteIndexConfig, profile RemoteIndexConfig) {
	if profile.URL != "" {
		target.URL = profile.URL
	}
	if profile.Audience != "" {
		target.Audience = profile.Audience
	}
	if profile.TenantID != "" {
		target.TenantID = profile.TenantID
	}
}

func validateConfigFile(root string) []ConfigFinding {
	path := filepath.Join(root, ".cairn", "config.yaml")
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []ConfigFinding{{Path: ".cairn/config.yaml", Severity: "error", Message: err.Error()}}
	}
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	var findings []ConfigFinding
	seen := map[string]bool{}
	section := ""
	profile := ""
	podRemoteEnabled := false
	podRemoteFields := map[string]string{}
	defaultTypes := DefaultConfig().DocumentTypes
	managed := map[string]bool{}
	for index, raw := range lines {
		lineNo := index + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "\t") {
			findings = append(findings, configFinding("error", lineNo, "tabs are not supported in config YAML"))
		}
		if !strings.HasPrefix(raw, " ") {
			if strings.HasSuffix(line, ":") {
				name := strings.TrimSuffix(line, ":")
				seen[name] = true
				if isConfigSection(name) {
					section = name
					profile = ""
					continue
				}
				section = ""
				if name == "workspace_id" {
					findings = append(findings, configFinding("error", lineNo, "workspace_id is required"))
				} else if name == "schema_version" {
					findings = append(findings, configFinding("error", lineNo, "schema_version must be 1"))
				} else {
					findings = append(findings, configFinding("error", lineNo, "unknown empty config key "+name))
				}
				continue
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, configFinding("error", lineNo, "malformed top-level config entry"))
				section = ""
				continue
			}
			key = strings.TrimSpace(key)
			seen[key] = true
			section = ""
			profile = ""
			switch key {
			case "schema_version":
				if strings.TrimSpace(value) != "1" {
					findings = append(findings, configFinding("error", lineNo, "schema_version must be 1"))
				}
			case "workspace_id":
				if unquoteConfig(strings.TrimSpace(value)) == "" {
					findings = append(findings, configFinding("error", lineNo, "workspace_id is required"))
				}
			case "managed_folders", "document_types", "remote_sync", "remote_index", "profiles", "required_skills":
				findings = append(findings, configFinding("warning", lineNo, "section should be declared without an inline value"))
			}
			continue
		}
		if section == "profiles" {
			indent := leadingSpaces(raw)
			if indent == 2 && strings.HasSuffix(line, ":") {
				profile = strings.TrimSuffix(line, ":")
				if profile != "local" && profile != "pod-remote" {
					findings = append(findings, configFinding("warning", lineNo, "unknown profile "+profile))
				}
				continue
			}
			if indent < 4 || profile == "" {
				findings = append(findings, configFinding("error", lineNo, "profiles entries must be nested under a profile name"))
				continue
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, configFinding("error", lineNo, "profile entry is malformed"))
				continue
			}
			key = strings.TrimSpace(key)
			value = unquoteConfig(strings.TrimSpace(value))
			switch key {
			case "enabled":
				if value != "true" && value != "false" {
					findings = append(findings, configFinding("error", lineNo, "profile enabled must be true or false"))
				}
				if profile == "pod-remote" && value == "true" {
					podRemoteEnabled = true
				}
			case "provider", "account", "endpoint", "container", "prefix", "root", "auth_mode", "url", "audience", "tenant_id":
				if profile == "pod-remote" {
					podRemoteFields[key] = value
				}
				if key == "provider" && value != "" && value != "azure_blob" && value != "local_fs" {
					findings = append(findings, configFinding("warning", lineNo, "unknown remote_sync provider "+value))
				}
			default:
				findings = append(findings, configFinding("warning", lineNo, "unknown profile key "+key))
			}
			continue
		}
		switch section {
		case "managed_folders":
			if !strings.HasPrefix(line, "- ") {
				findings = append(findings, configFinding("error", lineNo, "managed_folders entries must use list syntax"))
				continue
			}
			folder := cleanConfigPath(unquoteConfig(strings.TrimSpace(strings.TrimPrefix(line, "- "))))
			if folder == "" {
				findings = append(findings, configFinding("error", lineNo, "managed folder path is invalid"))
				continue
			}
			managed[folder] = true
		case "document_types":
			key, value, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, configFinding("error", lineNo, "document type mapping is malformed"))
				continue
			}
			docType := strings.TrimSpace(key)
			folder := cleanConfigPath(unquoteConfig(strings.TrimSpace(value)))
			if folder == "" {
				findings = append(findings, configFinding("error", lineNo, "document type destination is invalid"))
				continue
			}
			if _, ok := defaultTypes[docType]; !ok {
				findings = append(findings, configFinding("warning", lineNo, "unknown document type mapping "+docType))
			}
			managed[folder] = true
		case "remote_sync":
			key, value, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, configFinding("error", lineNo, "remote_sync entry is malformed"))
				continue
			}
			key = strings.TrimSpace(key)
			value = unquoteConfig(strings.TrimSpace(value))
			switch key {
			case "provider":
				if value != "" && value != "azure_blob" && value != "local_fs" {
					findings = append(findings, configFinding("warning", lineNo, "unknown remote_sync provider "+value))
				}
			case "account", "endpoint", "container", "prefix", "root", "auth_mode":
			default:
				findings = append(findings, configFinding("warning", lineNo, "unknown remote_sync key "+key))
			}
		case "remote_index":
			key, _, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, configFinding("error", lineNo, "remote_index entry is malformed"))
				continue
			}
			key = strings.TrimSpace(key)
			switch key {
			case "url", "audience", "tenant_id":
			default:
				findings = append(findings, configFinding("warning", lineNo, "unknown remote_index key "+key))
			}
		case "":
			findings = append(findings, configFinding("error", lineNo, "indented config entry is outside a section"))
		}
	}
	if podRemoteEnabled {
		if podRemoteFields["provider"] == "" {
			findings = append(findings, configFinding("error", 0, "pod-remote provider is required when enabled"))
		}
		if podRemoteFields["provider"] == "local_fs" {
			if podRemoteFields["root"] == "" {
				findings = append(findings, configFinding("error", 0, "pod-remote root is required when local_fs is enabled"))
			}
		} else if podRemoteFields["account"] == "" && podRemoteFields["endpoint"] == "" {
			findings = append(findings, configFinding("error", 0, "pod-remote account or endpoint is required when enabled"))
		}
		if podRemoteFields["provider"] != "local_fs" && podRemoteFields["container"] == "" {
			findings = append(findings, configFinding("error", 0, "pod-remote container is required when enabled"))
		}
	}
	for _, required := range []string{"schema_version", "workspace_id", "managed_folders", "document_types"} {
		if !seen[required] {
			findings = append(findings, ConfigFinding{Path: ".cairn/config.yaml", Severity: "error", Message: "missing required config key " + required})
		}
	}
	if len(managed) == 0 && seen["managed_folders"] {
		findings = append(findings, ConfigFinding{Path: ".cairn/config.yaml", Severity: "error", Message: "managed_folders must include at least one valid folder"})
	}
	return findings
}

func isConfigSection(name string) bool {
	switch name {
	case "managed_folders", "document_types", "remote_sync", "remote_index", "profiles", "required_skills":
		return true
	default:
		return false
	}
}

func leadingSpaces(value string) int {
	count := 0
	for _, r := range value {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}

func configFinding(severity string, line int, message string) ConfigFinding {
	return ConfigFinding{
		Path:     ".cairn/config.yaml",
		Severity: severity,
		Message:  message + " at line " + strconv.Itoa(line),
	}
}

func validateSchemaFiles(root string) []ConfigFinding {
	var findings []ConfigFinding
	schemaRoot := filepath.Join(root, ".cairn", "schemas")
	entries, err := os.ReadDir(schemaRoot)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []ConfigFinding{{Path: ".cairn/schemas", Severity: "error", Message: err.Error()}}
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".yaml") {
			continue
		}
		rel := filepath.ToSlash(filepath.Join(".cairn", "schemas", entry.Name()))
		content, err := os.ReadFile(filepath.Join(schemaRoot, entry.Name()))
		if err != nil {
			findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: err.Error()})
			continue
		}
		findings = append(findings, validateSchemaContent(rel, string(content))...)
	}
	return findings
}

func validateSchemaContent(rel string, content string) []ConfigFinding {
	var findings []ConfigFinding
	section := ""
	required := map[string]bool{}
	seenVersion := false
	seenRequired := false
	for index, raw := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		lineNo := index + 1
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(raw, " ") {
			if strings.HasSuffix(line, ":") {
				section = strings.TrimSuffix(line, ":")
				if section == "required_fields" {
					seenRequired = true
				}
				continue
			}
			key, value, ok := strings.Cut(line, ":")
			if !ok || strings.TrimSpace(key) == "" {
				findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "malformed schema entry at line " + strconv.Itoa(lineNo)})
				section = ""
				continue
			}
			section = ""
			if strings.TrimSpace(key) == "schema_version" {
				seenVersion = true
				if strings.TrimSpace(value) != "1" {
					findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "schema_version must be 1 at line " + strconv.Itoa(lineNo)})
				}
			}
			continue
		}
		if section == "required_fields" {
			if !strings.HasPrefix(line, "- ") {
				findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "required_fields entries must use list syntax at line " + strconv.Itoa(lineNo)})
				continue
			}
			field := unquoteConfig(strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			if field != "" {
				required[field] = true
			}
		}
	}
	if !seenVersion {
		findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "missing required schema_version"})
	}
	if !seenRequired {
		findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "missing required_fields"})
	}
	for _, field := range coreSchemaRequiredFields {
		if !required[field] {
			findings = append(findings, ConfigFinding{Path: rel, Severity: "error", Message: "required_fields must include Cairn core field " + field})
		}
	}
	return findings
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
