package syncstate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

type ObjectPublisher interface {
	WriteManifest(ctx context.Context, manifest Manifest) error
	WriteObject(ctx context.Context, path string, content []byte) error
	DeleteObject(ctx context.Context, path string) error
}

type PushOptions struct {
	WorkspaceID string
	Now         func() time.Time
}

func ApplyPush(ctx context.Context, root string, status Status, store ObjectPublisher, opts PushOptions) (Plan, error) {
	if store == nil {
		return Plan{}, errors.New("remote store is required for sync push")
	}
	plan := PlanFromStatus(status)
	if plan.Direction != PlanDirectionPush || !plan.Safe {
		return plan, fmt.Errorf("sync push refused for %s plan", plan.Direction)
	}

	localManifest, err := Generate(root, GenerateOptions{WorkspaceID: opts.WorkspaceID, Now: opts.Now})
	if err != nil {
		return plan, err
	}
	localManifest = statusComparableManifest(localManifest)
	if err := validateLocalManifest(root, localManifest); err != nil {
		return plan, err
	}
	localByPath := entriesByPath(localManifest.Entries)
	for _, change := range plan.Changes {
		if err := ctx.Err(); err != nil {
			return plan, err
		}
		switch change.Type {
		case ChangeCreate, ChangeEdit:
			entry, ok := localByPath[change.Path]
			if !ok {
				return plan, fmt.Errorf("local manifest is missing %s", change.Path)
			}
			if err := pushWrite(ctx, root, store, entry); err != nil {
				return plan, err
			}
		case ChangeMove, ChangeArchive:
			entry, ok := localByPath[change.Path]
			if !ok {
				return plan, fmt.Errorf("local manifest is missing %s", change.Path)
			}
			if err := pushWrite(ctx, root, store, entry); err != nil {
				return plan, err
			}
			if change.PreviousPath != "" && change.PreviousPath != change.Path {
				if err := store.DeleteObject(ctx, change.PreviousPath); err != nil {
					return plan, err
				}
			}
		case ChangeDelete:
			if err := store.DeleteObject(ctx, change.Path); err != nil {
				return plan, err
			}
		default:
			return plan, fmt.Errorf("unsupported push change type %q", change.Type)
		}
	}

	if err := store.WriteManifest(ctx, localManifest); err != nil {
		return plan, err
	}
	localHash, err := Hash(localManifest)
	if err != nil {
		return plan, err
	}
	now := time.Now().UTC()
	if opts.Now != nil {
		now = opts.Now().UTC()
	}
	return plan, Save(root, State{
		LastRemoteManifestHash: localHash,
		LastSyncAt:             now,
		Entries:                localManifest.Entries,
	})
}

func pushWrite(ctx context.Context, root string, store ObjectPublisher, entry Entry) error {
	absolutePath, err := workspacePath(root, entry.Path)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(absolutePath)
	if err != nil {
		return err
	}
	if entry.Hash != "" && hashBytes(content) != entry.Hash {
		return fmt.Errorf("local object %s hash does not match manifest", entry.Path)
	}
	return store.WriteObject(ctx, entry.Path, content)
}
