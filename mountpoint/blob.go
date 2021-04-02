package mountpoint

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type blobMount struct {
	*abstractMount

	BlobID string
}

func NewBlobMount(blobID string, client *menmos.Client, fs fs.Info) MountPoint {
	return &blobMount{
		abstractMount: &abstractMount{client, fs, newPathCache()},
		BlobID:        blobID,
	}
}

func (m *blobMount) ListEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	if pathSegment == "" || pathSegment == "." {
		return m.getEntriesFromQuery(payload.NewStructuredQuery(payload.NewExpression().AndParent(m.BlobID)), fullpath)
	}

	if err := m.ensurePathInCache(pathSegment); err != nil {
		return nil, err
	}

	targetDirBlobID, ok := m.cache.GetBlobID(pathSegment)
	if !ok {
		return nil, errors.New("cache walkback failed: unknown directory")
	}

	return m.getEntriesFromQuery(payload.NewStructuredQuery(payload.NewExpression().AndParent(targetDirBlobID)), fullpath)
}

func (m *blobMount) ResolveBlobDirectory(path string) (*entry.DirectoryBlobEntry, bool) {
	if path == "" {
		meta, err := m.client.GetMetadata(m.BlobID)
		if err != nil {
			return nil, false
		}

		return entry.NewDirectory(m.BlobID, meta, "", m.client, m.fs), true
	}

	parentDir := filepath.Dir(path)
	base := filepath.Base(path)
	entries, err := m.ListEntries(context.Background(), parentDir, parentDir)
	if err != nil {
		// TODO: Log
		return nil, false
	}

	for _, dirEntry := range entries {
		if filepath.Base(dirEntry.Remote()) != base {
			continue
		}

		if blobEntry, ok := dirEntry.(*entry.DirectoryBlobEntry); ok {
			return blobEntry, true
		}
	}

	return nil, false
}

func (m *blobMount) ResolveBlobFile(path string) (*entry.FileBlobEntry, bool) {
	fs.Infof(nil, "Blob Mount resolving blob file: %s", path)
	parentDir := filepath.Dir(path)
	base := filepath.Base(path)
	entries, err := m.ListEntries(context.Background(), parentDir, parentDir)
	if err != nil {
		// TODO: Log
		return nil, false
	}

	for _, dirEntry := range entries {
		if filepath.Base(dirEntry.Remote()) != base {
			continue
		}

		if blobEntry, ok := dirEntry.(*entry.FileBlobEntry); ok {
			return blobEntry, true
		}
	}

	return nil, false
}
