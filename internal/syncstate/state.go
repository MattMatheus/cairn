package syncstate

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const StateVersion = 1

type State struct {
	StateVersion           int       `json:"state_version"`
	LastRemoteManifestHash string    `json:"last_remote_manifest_hash"`
	LastSyncAt             time.Time `json:"last_sync_at"`
	Entries                []Entry   `json:"entries"`
}

func Load(root string) (State, error) {
	content, err := os.ReadFile(statePath(root))
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(content, &state); err != nil {
		return State{}, err
	}
	normalizeEntries(state.Entries)
	return state, nil
}

func Save(root string, state State) error {
	state.StateVersion = StateVersion
	state.LastSyncAt = state.LastSyncAt.UTC().Truncate(time.Second)
	normalizeEntries(state.Entries)
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	if err := os.MkdirAll(filepath.Join(root, ".cairn"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(statePath(root), content, 0o644)
}

func statePath(root string) string {
	return filepath.Join(root, ".cairn", "sync-state.json")
}
