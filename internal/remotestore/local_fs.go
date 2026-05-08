package remotestore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/syncstate"
)

type LocalFSStore struct {
	Root   string
	Prefix string
}

func NewLocalFSStore(root string, prefix string) (*LocalFSStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("local_fs root is required")
	}
	return &LocalFSStore{Root: root, Prefix: prefix}, nil
}

func (s *LocalFSStore) ReadManifest(ctx context.Context) (syncstate.Manifest, bool, error) {
	content, ok, err := s.ReadObject(ctx, RemoteManifestPath)
	if err != nil || !ok {
		return syncstate.Manifest{}, ok, err
	}
	var manifest syncstate.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return syncstate.Manifest{}, false, err
	}
	return manifest, true, nil
}

func (s *LocalFSStore) WriteManifest(ctx context.Context, manifest syncstate.Manifest) error {
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return s.WriteObject(ctx, RemoteManifestPath, append(content, '\n'))
}

func (s *LocalFSStore) ReadObject(ctx context.Context, path string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	absolutePath, err := s.objectPath(path)
	if err != nil {
		return nil, false, err
	}
	content, err := os.ReadFile(absolutePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	return content, err == nil, err
}

func (s *LocalFSStore) WriteObject(ctx context.Context, path string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	absolutePath, err := s.objectPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(absolutePath, content, 0o644)
}

func (s *LocalFSStore) DeleteObject(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	absolutePath, err := s.objectPath(path)
	if err != nil {
		return err
	}
	if err := os.Remove(absolutePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *LocalFSStore) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rootPrefix := JoinPrefix(s.Prefix, prefix)
	walkRoot, err := s.rawObjectPath(rootPrefix)
	if err != nil {
		return nil, err
	}
	var objects []ObjectInfo
	err = filepath.WalkDir(walkRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if errors.Is(walkErr, os.ErrNotExist) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(s.Root, path)
		if err != nil {
			return err
		}
		objects = append(objects, ObjectInfo{
			Path: StripPrefix(s.Prefix, filepath.ToSlash(rel)),
			Size: info.Size(),
		})
		return nil
	})
	return objects, err
}

func (s *LocalFSStore) objectPath(path string) (string, error) {
	return s.rawObjectPath(JoinPrefix(s.Prefix, path))
}

func (s *LocalFSStore) rawObjectPath(objectName string) (string, error) {
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return "", err
	}
	absolutePath, err := filepath.Abs(filepath.Join(root, filepath.FromSlash(objectName)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("unsafe local_fs object path")
	}
	return absolutePath, nil
}
