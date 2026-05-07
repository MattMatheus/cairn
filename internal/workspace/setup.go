package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SetupLocalSyncOptions struct {
	WorkspaceID string
	RemoteRoot  string
	Force       bool
}

type SetupLocalSyncResult struct {
	WorkspaceID  string
	ConfigPath   string
	RemoteRoot   string
	InitCreated  []string
	InitExisting []string
}

type SetupAzureSyncOptions struct {
	WorkspaceID string
	Account     string
	Endpoint    string
	Container   string
	Prefix      string
	Force       bool
}

type SetupAzureSyncResult struct {
	WorkspaceID  string
	ConfigPath   string
	Account      string
	Endpoint     string
	Container    string
	Prefix       string
	InitCreated  []string
	InitExisting []string
}

func SetupLocalSync(root string, opts SetupLocalSyncOptions) (SetupLocalSyncResult, error) {
	remoteRoot := strings.TrimSpace(opts.RemoteRoot)
	if remoteRoot == "" {
		return SetupLocalSyncResult{}, errors.New("local sync setup requires --remote-root")
	}
	initResult, err := Init(root, InitOptions{WorkspaceID: opts.WorkspaceID, Force: opts.Force})
	if err != nil {
		return SetupLocalSyncResult{}, err
	}
	configPath := filepath.Join(root, ".cairn", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return SetupLocalSyncResult{}, err
	}
	updated, err := ensureLocalSyncConfig(string(content), remoteRoot)
	if err != nil {
		return SetupLocalSyncResult{}, err
	}
	if updated != string(content) {
		if err := os.WriteFile(configPath, []byte(updated), 0o644); err != nil {
			return SetupLocalSyncResult{}, err
		}
	}
	return SetupLocalSyncResult{
		WorkspaceID:  initResult.WorkspaceID,
		ConfigPath:   ".cairn/config.yaml",
		RemoteRoot:   remoteRoot,
		InitCreated:  initResult.Created,
		InitExisting: initResult.Existing,
	}, nil
}

func SetupAzureSync(root string, opts SetupAzureSyncOptions) (SetupAzureSyncResult, error) {
	account := strings.TrimSpace(opts.Account)
	endpoint := strings.TrimSpace(opts.Endpoint)
	container := strings.TrimSpace(opts.Container)
	prefix := strings.Trim(strings.TrimSpace(opts.Prefix), "/")
	if account == "" && endpoint == "" {
		return SetupAzureSyncResult{}, errors.New("azure sync setup requires --account or --endpoint")
	}
	if container == "" {
		return SetupAzureSyncResult{}, errors.New("azure sync setup requires --container")
	}
	initResult, err := Init(root, InitOptions{WorkspaceID: opts.WorkspaceID, Force: opts.Force})
	if err != nil {
		return SetupAzureSyncResult{}, err
	}
	configPath := filepath.Join(root, ".cairn", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return SetupAzureSyncResult{}, err
	}
	updated, err := ensureAzureSyncConfig(string(content), account, endpoint, container, prefix)
	if err != nil {
		return SetupAzureSyncResult{}, err
	}
	if updated != string(content) {
		if err := os.WriteFile(configPath, []byte(updated), 0o644); err != nil {
			return SetupAzureSyncResult{}, err
		}
	}
	return SetupAzureSyncResult{
		WorkspaceID:  initResult.WorkspaceID,
		ConfigPath:   ".cairn/config.yaml",
		Account:      account,
		Endpoint:     endpoint,
		Container:    container,
		Prefix:       prefix,
		InitCreated:  initResult.Created,
		InitExisting: initResult.Existing,
	}, nil
}

func ensureLocalSyncConfig(content string, remoteRoot string) (string, error) {
	if strings.Contains(content, "\nremote_sync:") || strings.HasPrefix(content, "remote_sync:") {
		if strings.Contains(content, "provider: local_fs") && (strings.Contains(content, "root: "+quoteConfigValue(remoteRoot)) || strings.Contains(content, "root: "+remoteRoot)) {
			return content, nil
		}
		return "", errors.New("remote_sync is already configured; edit .cairn/config.yaml to change it")
	}
	trimmed := strings.TrimRight(content, "\n")
	return fmt.Sprintf("%s\nremote_sync:\n  provider: local_fs\n  root: %s\n", trimmed, quoteConfigValue(remoteRoot)), nil
}

func ensureAzureSyncConfig(content string, account string, endpoint string, container string, prefix string) (string, error) {
	if strings.Contains(content, "\nremote_sync:") || strings.HasPrefix(content, "remote_sync:") {
		prefixMatches := (prefix == "" && !strings.Contains(content, "\n  prefix:")) ||
			(prefix != "" && (strings.Contains(content, "prefix: "+quoteConfigValue(prefix)) || strings.Contains(content, "prefix: "+prefix)))
		if strings.Contains(content, "provider: azure_blob") &&
			(account == "" || strings.Contains(content, "account: "+quoteConfigValue(account)) || strings.Contains(content, "account: "+account)) &&
			(endpoint == "" || strings.Contains(content, "endpoint: "+quoteConfigValue(endpoint)) || strings.Contains(content, "endpoint: "+endpoint)) &&
			(strings.Contains(content, "container: "+quoteConfigValue(container)) || strings.Contains(content, "container: "+container)) &&
			prefixMatches {
			return content, nil
		}
		return "", errors.New("remote_sync is already configured; edit .cairn/config.yaml to change it")
	}
	var builder strings.Builder
	builder.WriteString(strings.TrimRight(content, "\n"))
	builder.WriteString("\nremote_sync:\n  provider: azure_blob\n")
	if account != "" {
		builder.WriteString("  account: ")
		builder.WriteString(quoteConfigValue(account))
		builder.WriteString("\n")
	}
	if endpoint != "" {
		builder.WriteString("  endpoint: ")
		builder.WriteString(quoteConfigValue(endpoint))
		builder.WriteString("\n")
	}
	builder.WriteString("  container: ")
	builder.WriteString(quoteConfigValue(container))
	builder.WriteString("\n")
	if prefix != "" {
		builder.WriteString("  prefix: ")
		builder.WriteString(quoteConfigValue(prefix))
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

func quoteConfigValue(value string) string {
	return strconv.Quote(value)
}
