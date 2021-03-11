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

type queryMount struct {
	Expression payload.Expression

	GroupByTags     bool
	GroupByMetaKeys []string

	client *menmos.Client
	fs     fs.Info

	cache *pathCache
}

func NewQueryMount(expression payload.Expression, groupByTags bool, groupByMetaKeys []string, client *menmos.Client, fs fs.Info) *queryMount {
	return &queryMount{
		Expression:      expression,
		GroupByTags:     groupByTags,
		GroupByMetaKeys: groupByMetaKeys,
		client:          client,
		fs:              fs,
		cache:           newPathCache(),
	}
}

func (m *queryMount) groupByKeysContains(key string) bool {
	for _, v := range m.GroupByMetaKeys {
		if v == key {
			return true
		}
	}
	return false
}

func (m *queryMount) getQueryChildrenMap(query *payload.Query) (map[string]string, error) {
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

func (m *queryMount) getEntriesFromQuery(query *payload.Query, fullpath string) (fs.DirEntries, error) {
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

func (m *queryMount) ensurePathInCache(pathSegment string) error {
	// Cache the blob IDs of all directories in the way of the path we're trying to reach.
	// TODO: Document this part some more because its confusing as hell.

	// Walk back the cache to find the lowest cached directory in the path we're looking for (if any).
	fmt.Printf("resolving blob IDS for '%s'\n", pathSegment)

	cachedSegment := pathSegment
	lowestCachedBlobID := ""
	for cachedSegment != string(os.PathSeparator) && cachedSegment != "." {
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

	fmt.Printf("cachedSegment='%s', uncachedSegment='%s', lowestCachedBlobID='%s'\n", cachedSegment, uncachedSegment, lowestCachedBlobID)

	for {
		cachedBlobEntries, err := m.getQueryChildrenMap(payload.NewStructuredQuery(payload.NewExpression().AndParent(lowestCachedBlobID)))
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

func (m *queryMount) listNestedEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	fmt.Println("Listing nested entries")
	if pathSegment == "" {
		entries := make([]fs.DirEntry, 0, len(m.GroupByMetaKeys)+1)
		entries = append(entries, &entry.VDirEntry{Name: "Tags", FullPath: path.Join(fullpath, "Tags")})
		for _, metaKey := range m.GroupByMetaKeys {
			entries = append(entries, &entry.VDirEntry{Name: metaKey, FullPath: path.Join(fullpath, metaKey)})
		}
		return entries, nil
	}

	splitted := strings.SplitN(pathSegment, string(os.PathSeparator), 2)
	head := splitted[0]

	rootQuery := payload.NewStructuredQuery(m.Expression).WithSize(1000) // TODO: Paging
	results, err := m.client.Query(rootQuery.WithFacets(true))
	if err != nil {
		return nil, err
	}

	facets := results.Facets
	if facets == nil {
		return nil, errors.New("no facets returned")
	}

	if head == "Tags" {
		tagMountMap := make(map[string]MountPoint)
		for tag := range facets.Tags {
			tagMountMap[tag] = NewQueryMount(m.Expression.AndTag(tag), false, []string{}, m.client, m.fs)
		}
		mount := &virtualMount{mounts: tagMountMap}

		tail := ""
		if len(splitted) == 2 {
			tail = splitted[1]
		}

		return mount.ListEntries(ctx, tail, fullpath)
	} else if m.groupByKeysContains(head) { // Head is a k/v key
		fmt.Println("HEAD is K/V: ", head)
		fmt.Println("splitted")
		kvMountMap := make(map[string]MountPoint)
		for value := range facets.Meta[head] {
			kvMountMap[value] = NewQueryMount(m.Expression.AndKeyValue(head, value), false, []string{}, m.client, m.fs)
		}
		mount := &virtualMount{mounts: kvMountMap}

		tail := ""
		if len(splitted) == 2 {
			tail = splitted[1]
		}

		return mount.ListEntries(ctx, tail, fullpath)
	}

	return nil, fs.ErrorDirNotFound
}

func (m *queryMount) listFlatEntries(pathSegment string, fullpath string) (fs.DirEntries, error) {
	rootQuery := payload.NewStructuredQuery(m.Expression).WithSize(1000) // TODO: Paging
	if pathSegment == "" || pathSegment == "." {
		return m.getEntriesFromQuery(rootQuery, fullpath)
	}

	// Pre-populate the cache with virtual files & directories.
	childrenMap, err := m.getQueryChildrenMap(rootQuery)
	if err != nil {
		return nil, err
	}
	for name, blobID := range childrenMap {
		m.cache.SetBlobID(name, blobID)
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

func (m *queryMount) ListEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	fmt.Printf("listing query entries for '%s'\n", pathSegment)

	shouldGroup := m.GroupByTags || len(m.GroupByMetaKeys) > 0

	if shouldGroup {
		return m.listNestedEntries(ctx, pathSegment, fullpath)
	}

	return m.listFlatEntries(pathSegment, fullpath)
}

func (m *queryMount) ResolveBlobDirectory(path string) (*entry.DirectoryBlobEntry, bool) {
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

func (m *queryMount) ResolveBlobFile(path string) (*entry.FileBlobEntry, bool) {
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
