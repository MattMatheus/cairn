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
}

type SetupLocalSyncResult struct {
	WorkspaceID  string
	ConfigPath   string
	RemoteRoot   string
	InitCreated  []string
	InitExisting []string
}

func SetupLocalSync(root string, opts SetupLocalSyncOptions) (SetupLocalSyncResult, error) {
	remoteRoot := strings.TrimSpace(opts.RemoteRoot)
	if remoteRoot == "" {
		return SetupLocalSyncResult{}, errors.New("local sync setup requires --remote-root")
	}
	initResult, err := Init(root, InitOptions{WorkspaceID: opts.WorkspaceID})
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

func quoteConfigValue(value string) string {
	return strconv.Quote(value)
}
