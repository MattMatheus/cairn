package syncstate

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"cairn/internal/mcpschema"
)

type StatusOptions struct {
	WorkspaceID        string
	RemoteManifestPath string
	RemoteManifest     *Manifest
	Now                func() time.Time
}

type Status struct {
	BaseManifestHash   string
	RemoteManifestHash string
	State              State
	LocalManifest      Manifest
	RemoteManifest     Manifest
	Comparison         Comparison
	RemoteAvailable    bool
}

func StatusReport(ctx context.Context, root string, opts StatusOptions) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	state, err := Load(root)
	if err != nil {
		return Status{}, err
	}
	local, err := Generate(root, GenerateOptions{WorkspaceID: opts.WorkspaceID, Now: opts.Now})
	if err != nil {
		return Status{}, err
	}
	local = statusComparableManifest(local)
	var remote Manifest
	remoteAvailable := false
	if opts.RemoteManifest != nil {
		remote = *opts.RemoteManifest
		remoteAvailable = true
	} else {
		var err error
		remote, remoteAvailable, err = loadRemoteManifest(remoteManifestPath(root, opts.RemoteManifestPath))
		if err != nil {
			return Status{}, err
		}
	}
	remote = statusComparableManifest(remote)

	base := Manifest{
		ManifestVersion: ManifestVersion,
		WorkspaceID:     opts.WorkspaceID,
		Entries:         statusComparableEntries(state.Entries),
	}
	if !remoteAvailable {
		remote = base
	}
	comparison := Compare(base, local, remote)
	remoteHash := ""
	if remoteAvailable {
		remoteHash, err = Hash(remote)
		if err != nil {
			return Status{}, err
		}
	}
	return Status{
		BaseManifestHash:   state.LastRemoteManifestHash,
		RemoteManifestHash: remoteHash,
		State:              state,
		LocalManifest:      local,
		RemoteManifest:     remote,
		Comparison:         comparison,
		RemoteAvailable:    remoteAvailable,
	}, nil
}

func statusComparableManifest(manifest Manifest) Manifest {
	manifest.Entries = statusComparableEntries(manifest.Entries)
	normalizeEntries(manifest.Entries)
	return manifest
}

func statusComparableEntries(entries []Entry) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		switch entry.Path {
		case ".cairn/sync-state.json", ".cairn/remote-manifest.json":
			continue
		default:
			out = append(out, entry)
		}
	}
	return out
}

func Envelope(status Status) mcpschema.Envelope[mcpschema.SyncStatusData] {
	envelope := mcpschema.Envelope[mcpschema.SyncStatusData]{
		OK: true,
		Data: mcpschema.SyncStatusData{
			LocalChanges:  schemaChanges(status.Comparison.LocalChanges),
			RemoteChanges: schemaChanges(status.Comparison.RemoteChanges),
			Conflicts:     schemaConflicts(status.Comparison.Conflicts),
			Diverged:      status.Comparison.Diverged,
			LastSyncAt:    status.State.LastSyncAt,
			BaseHash:      status.BaseManifestHash,
			RemoteHash:    status.RemoteManifestHash,
		},
	}
	if !status.RemoteAvailable {
		envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
			Code:    mcpschema.WarningRemoteService,
			Message: "remote manifest is unavailable in local profile",
			Path:    ".cairn/remote-manifest.json",
		})
		envelope.Unavailable = append(envelope.Unavailable, mcpschema.UnavailableMode{
			Mode:      "remote_manifest",
			Reason:    mcpschema.WarningRemoteService,
			Message:   "remote manifest fixture is not available",
			Retryable: false,
		})
	}
	if status.Comparison.Diverged {
		envelope.OK = false
		envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
			Code:    mcpschema.WarningSyncDivergence,
			Message: "local and remote changes diverged; sync mutation must be refused",
		})
		envelope.NextSteps = append(envelope.NextSteps,
			mcpschema.NextStep{Action: "review_conflicts", Label: "Review conflicting paths", Reason: "Manual reconciliation is required before pull or push."},
			mcpschema.NextStep{Action: string(mcpschema.ToolSyncStatus), Label: "Run sync status again", Reason: "Retry after reconciling local or remote changes."},
		)
		return envelope
	}
	if len(status.Comparison.LocalChanges) > 0 {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolSyncPush),
			Label:  "Push local changes when remote sync is configured",
			Reason: "Only local changes were detected.",
		})
	}
	if len(status.Comparison.RemoteChanges) > 0 {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolSyncPull),
			Label:  "Pull remote changes when remote sync is configured",
			Reason: "Only remote changes were detected.",
		})
	}
	if len(envelope.NextSteps) == 0 {
		envelope.NextSteps = append(envelope.NextSteps, mcpschema.NextStep{
			Action: string(mcpschema.ToolValidateWorkspace),
			Label:  "Validate workspace",
			Reason: "No sync changes were detected.",
		})
	}
	return envelope
}

func loadRemoteManifest(path string) (Manifest, bool, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Manifest{}, false, nil
	}
	if err != nil {
		return Manifest{}, false, err
	}
	var manifest Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return Manifest{}, false, err
	}
	normalizeEntries(manifest.Entries)
	return manifest, true, nil
}

func remoteManifestPath(root string, path string) string {
	if path != "" {
		return path
	}
	return filepath.Join(root, ".cairn", "remote-manifest.json")
}

func schemaChanges(changes []Change) []mcpschema.SyncChange {
	out := make([]mcpschema.SyncChange, 0, len(changes))
	for _, change := range changes {
		out = append(out, mcpschema.SyncChange{
			Type:         string(change.Type),
			Path:         change.Path,
			PreviousPath: change.PreviousPath,
			DocumentID:   change.DocumentID,
		})
	}
	return out
}

func schemaConflicts(conflicts []Conflict) []mcpschema.SyncConflict {
	out := make([]mcpschema.SyncConflict, 0, len(conflicts))
	for _, conflict := range conflicts {
		local := schemaChanges([]Change{conflict.Local})
		remote := schemaChanges([]Change{conflict.Remote})
		out = append(out, mcpschema.SyncConflict{Local: local[0], Remote: remote[0]})
	}
	return out
}
