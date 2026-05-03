package remotestore

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"cairn/internal/syncstate"
)

type MemoryStore struct {
	Objects map[string][]byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{Objects: map[string][]byte{}}
}

func (s *MemoryStore) ReadManifest(ctx context.Context) (syncstate.Manifest, bool, error) {
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

func (s *MemoryStore) WriteManifest(ctx context.Context, manifest syncstate.Manifest) error {
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return s.WriteObject(ctx, RemoteManifestPath, append(content, '\n'))
}

func (s *MemoryStore) ReadObject(ctx context.Context, path string) ([]byte, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	s.ensure()
	content, ok := s.Objects[CleanPath(path)]
	if !ok {
		return nil, false, nil
	}
	copyContent := append([]byte(nil), content...)
	return copyContent, true, nil
}

func (s *MemoryStore) WriteObject(ctx context.Context, path string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.ensure()
	s.Objects[CleanPath(path)] = append([]byte(nil), content...)
	return nil
}

func (s *MemoryStore) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.ensure()
	prefix = CleanPath(prefix)
	var objects []ObjectInfo
	for path, content := range s.Objects {
		if prefix != "" && path != prefix && !strings.HasPrefix(path, prefix+"/") {
			continue
		}
		objects = append(objects, ObjectInfo{Path: path, Size: int64(len(content))})
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].Path < objects[j].Path
	})
	return objects, nil
}

func (s *MemoryStore) ensure() {
	if s.Objects == nil {
		s.Objects = map[string][]byte{}
	}
}
