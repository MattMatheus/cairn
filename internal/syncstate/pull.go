package syncstate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ObjectReader interface {
	ReadObject(ctx context.Context, path string) ([]byte, bool, error)
}

type PullOptions struct {
	Now func() time.Time
}

type pulledObject struct {
	entry   Entry
	content []byte
}

func ApplyPull(ctx context.Context, root string, status Status, store ObjectReader, opts PullOptions) (Plan, error) {
	if store == nil {
		return Plan{}, errors.New("remote store is required for sync pull")
	}
	plan := PlanFromStatus(status)
	if plan.Direction != PlanDirectionPull || !plan.Safe {
		return plan, fmt.Errorf("sync pull refused for %s plan", plan.Direction)
	}

	remoteByPath := entriesByPath(status.RemoteManifest.Entries)
	pulled := map[string]pulledObject{}
	for _, change := range plan.Changes {
		if err := ctx.Err(); err != nil {
			return plan, err
		}
		if change.Type == ChangeDelete {
			continue
		}
		entry, ok := remoteByPath[change.Path]
		if !ok {
			return plan, fmt.Errorf("remote manifest is missing %s", change.Path)
		}
		content, err := readAndValidateRemoteObject(ctx, root, store, entry)
		if err != nil {
			return plan, err
		}
		pulled[change.Path] = pulledObject{entry: entry, content: content}
	}

	for _, change := range plan.Changes {
		if err := ctx.Err(); err != nil {
			return plan, err
		}
		switch change.Type {
		case ChangeCreate, ChangeEdit:
			object, ok := pulled[change.Path]
			if !ok {
				return plan, fmt.Errorf("remote object %s was not staged", change.Path)
			}
			if err := writePulledObject(root, object); err != nil {
				return plan, err
			}
		case ChangeMove, ChangeArchive:
			object, ok := pulled[change.Path]
			if !ok {
				return plan, fmt.Errorf("remote object %s was not staged", change.Path)
			}
			if err := writePulledObject(root, object); err != nil {
				return plan, err
			}
			if change.PreviousPath != "" && change.PreviousPath != change.Path {
				if err := removeWorkspacePath(root, change.PreviousPath); err != nil {
					return plan, err
				}
			}
		case ChangeDelete:
			if err := removeWorkspacePath(root, change.Path); err != nil {
				return plan, err
			}
		default:
			return plan, fmt.Errorf("unsupported pull change type %q", change.Type)
		}
	}

	remoteHash, err := Hash(status.RemoteManifest)
	if err != nil {
		return plan, err
	}
	now := time.Now().UTC()
	if opts.Now != nil {
		now = opts.Now().UTC()
	}
	return plan, Save(root, State{
		LastRemoteManifestHash: remoteHash,
		LastSyncAt:             now,
		Entries:                status.RemoteManifest.Entries,
	})
}

func entriesByPath(entries []Entry) map[string]Entry {
	out := make(map[string]Entry, len(entries))
	for _, entry := range entries {
		out[entry.Path] = entry
	}
	return out
}

func readAndValidateRemoteObject(ctx context.Context, root string, store ObjectReader, entry Entry) ([]byte, error) {
	content, ok, err := store.ReadObject(ctx, entry.Path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("remote object %s is missing", entry.Path)
	}
	if entry.Hash != "" && hashBytes(content) != entry.Hash {
		return nil, fmt.Errorf("remote object %s hash does not match manifest", entry.Path)
	}
	if err := validateRemoteContent(root, entry.Path, content); err != nil {
		return nil, err
	}
	return content, nil
}

func writePulledObject(root string, object pulledObject) error {
	absolutePath, err := workspacePath(root, object.entry.Path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absolutePath, object.content, 0o644)
}

func removeWorkspacePath(root string, relativePath string) error {
	absolutePath, err := workspacePath(root, relativePath)
	if err != nil {
		return err
	}
	if err := os.Remove(absolutePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func workspacePath(root string, relativePath string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || len(clean) >= 3 && clean[:3] == "../" {
		return "", fmt.Errorf("unsafe workspace path %q", relativePath)
	}
	return filepath.Join(root, clean), nil
}

func hashBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
