package mountpoint

import (
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type abstractMount struct {
	client *menmos.Client
	fs     fs.Info

	cache *pathCache
}

func (m *abstractMount) getQueryChildrenMap(query *payload.Query) (map[string]string, error) {
	results, err := getFullQueryResults(query, m.client)
	if err != nil {
		return nil, err
	}

	hitMap := make(map[string]string)

	for _, hit := range results.Hits {
		hitMap[hit.Metadata.Name] = hit.ID
	}

	return hitMap, nil
}

func (m *abstractMount) getEntriesFromQuery(query *payload.Query, fullpath string) (fs.DirEntries, error) {
	results, err := getFullQueryResults(query, m.client)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	entries := make([]fs.DirEntry, results.Count, results.Count)

	for i, hit := range results.Hits {
		if hit.Metadata.BlobType == "File" {
			fs.Infof(nil, "file entry for blob: %s", hit.ID)
			entries[i] = entry.NewFile(hit.ID, hit.Metadata, path.Join(fullpath, hit.Metadata.Name), m.client, m.fs)
		} else {
			fs.Infof(nil, "dir entry for blob: %s", hit.ID)
			entries[i] = entry.NewDirectory(hit.ID, hit.Metadata, path.Join(fullpath, hit.Metadata.Name), m.client, m.fs)
		}
	}

	return entries, nil
}

func (m *abstractMount) ensurePathInCache(pathSegment string) error {
	// Cache the blob IDs of all directories in the way of the path we're trying to reach.
	// TODO: Document this part some more because its confusing as hell.

	// Walk back the cache to find the lowest cached directory in the path we're looking for (if any).
	fs.Infof(nil, "resolving blob IDS for '%s'", pathSegment)

	cachedSegment := pathSegment
	lowestCachedBlobID := ""
	for cachedSegment != "/" && cachedSegment != "." {
		blobID, ok := m.cache.GetBlobID(cachedSegment)
		if ok {
			lowestCachedBlobID = blobID
			break
		}
		cachedSegment = filepath.Dir(cachedSegment)
	}

	if cachedSegment == "." || cachedSegment == "/" {
		cachedSegment = ""
	}

	if lowestCachedBlobID == "" {
		return errors.New("no base Blob ID found")
	}

	uncachedSegment := strings.TrimPrefix(pathSegment, cachedSegment)

	fs.Infof(nil, "cachedSegment='%s', uncachedSegment='%s', lowestCachedBlobID='%s'\n", cachedSegment, uncachedSegment, lowestCachedBlobID)

	for {
		cachedBlobEntries, err := m.getQueryChildrenMap(payload.NewStructuredQuery(payload.NewExpression().AndParent(lowestCachedBlobID)))
		if err != nil {
			return err
		}

		splitted := strings.SplitN(uncachedSegment, "/", 2)
		if splitted[0] == "/" || splitted[0] == "." || splitted[0] == "" {
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
