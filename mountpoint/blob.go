package mountpoint

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type blobMount struct {
	BlobID string

	client *menmos.Client
	fs     fs.Info

	cache *pathCache
}

func NewBlobMount(blobID string, client *menmos.Client, fs fs.Info) MountPoint {
	return &blobMount{
		BlobID: blobID,
		client: client,
		fs:     fs,
		cache:  newPathCache(),
	}
}

func (m *blobMount) getBlobChildrenMap(blobID string) (map[string]string, error) {
	query := payload.NewStructuredQuery(payload.NewExpression().AndParent(blobID))
	results, err := m.client.Query(query)
	if err != nil {
		return nil, err
	}

	hitMap := make(map[string]string)

	for _, hit := range results.Hits {
		hitMap[hit.Metadata.Name] = hit.ID
	}

	return hitMap, nil
}

func (m *blobMount) getBlobChildrenEntries(blobID string, fullpath string) (fs.DirEntries, error) {
	query := payload.NewStructuredQuery(payload.NewExpression().AndParent(blobID))
	results, err := m.client.Query(query)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	entries := make([]fs.DirEntry, results.Count, results.Count)

	for i, hit := range results.Hits {
		if hit.Metadata.BlobType == "File" {
			fmt.Println("file entry for blob: ", hit.ID)
			entries[i] = entry.NewFile(hit.ID, hit.Metadata, path.Join(fullpath, hit.Metadata.Name), m.client, m.fs)
		} else {
			fmt.Println("dir entry for blob: ", hit.ID)
			entries[i] = entry.NewDirectory(hit.ID, hit.Metadata, path.Join(fullpath, hit.Metadata.Name), m.client, m.fs)
		}
	}

	return entries, nil
}

func (m *blobMount) ensurePathInCache(pathSegment string) error {
	// Cache the blob IDs of all directories in the way of the path we're trying to reach.
	// TODO: Document this part some more because its confusing as hell.

	// Walk back the cache to find the lowest cached directory in the path we're looking for (if any).
	fmt.Printf("resolving blob IDS for '%s'\n", pathSegment)
	fmt.Printf("m.BlobID='%s'\n", m.BlobID)

	cachedSegment := pathSegment
	lowestCachedBlobID := m.BlobID
	for cachedSegment != string(os.PathSeparator) && cachedSegment != "." {
		blobID, ok := m.cache.GetBlobID(cachedSegment)
		if ok {
			lowestCachedBlobID = blobID
			break
		}
		cachedSegment = filepath.Dir(cachedSegment)
	}

	if cachedSegment == "." || cachedSegment == string(os.PathSeparator) {
		cachedSegment = ""
	}

	uncachedSegment := strings.TrimPrefix(strings.TrimPrefix(pathSegment, cachedSegment), string(os.PathSeparator))

	fmt.Printf("cachedSegment='%s', uncachedSegment='%s', lowestCachedBlobID='%s'\n", cachedSegment, uncachedSegment, lowestCachedBlobID)

	for {
		cachedBlobEntries, err := m.getBlobChildrenMap(lowestCachedBlobID)
		if err != nil {
			return err
		}

		splitted := strings.SplitN(uncachedSegment, string(os.PathSeparator), 2)
		if splitted[0] == string(os.PathSeparator) || splitted[0] == "." || splitted[0] == "" {
			// We reached the end.
			break
		}

		entryName := splitted[0]
		if entryBlobID, ok := cachedBlobEntries[entryName]; ok {
			cachedSegment = filepath.Join(cachedSegment, entryName)
			m.cache.SetBlobID(cachedSegment, entryBlobID)
		} else {
			return fs.ErrorDirNotFound
		}

		if len(splitted) == 1 {
			break
		}

		uncachedSegment = splitted[1]
	}

	return nil
}

func (m *blobMount) ListEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	if pathSegment == "" || pathSegment == "." {
		return m.getBlobChildrenEntries(m.BlobID, fullpath)
	}

	if err := m.ensurePathInCache(pathSegment); err != nil {
		return nil, err
	}

	targetDirBlobID, ok := m.cache.GetBlobID(pathSegment)
	if !ok {
		return nil, errors.New("cache walkback failed: unknown directory")
	}

	return m.getBlobChildrenEntries(targetDirBlobID, fullpath)
}

func (m *blobMount) ResolveBlobDirectory(path string) (*entry.DirectoryBlobEntry, bool) {
	if path == "" {
		// TODO: Get the metadata of "self"
		// TODO: This will explode spectacularly if this entry is actually used - figure out what's needed so it doesn't
		return &entry.DirectoryBlobEntry{BlobEntry: entry.BlobEntry{BlobID: m.BlobID}}, true
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
	fmt.Println("Blob Mount resolving blob file: ", path)
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
