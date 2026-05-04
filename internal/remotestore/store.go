package remotestore

import (
	"context"

	"cairn/internal/syncstate"
)

const RemoteManifestPath = ".cairn/remote-manifest.json"

type Store interface {
	ReadManifest(ctx context.Context) (syncstate.Manifest, bool, error)
	WriteManifest(ctx context.Context, manifest syncstate.Manifest) error
	ReadObject(ctx context.Context, path string) ([]byte, bool, error)
	WriteObject(ctx context.Context, path string, content []byte) error
	DeleteObject(ctx context.Context, path string) error
	ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error)
}

type ObjectInfo struct {
	Path string
	Size int64
}
