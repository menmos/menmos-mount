package mountpoint

import (
	"context"
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type queryMount struct {
	*abstractMount

	Expression payload.Expression

	GroupByTags     bool
	GroupByMetaKeys []string
}

func NewQueryMount(expression payload.Expression, groupByTags bool, groupByMetaKeys []string, client *menmos.Client, fs fs.Info) *queryMount {
	return &queryMount{
		abstractMount: &abstractMount{
			client,
			fs,
			newPathCache(),
		},
		Expression:      expression,
		GroupByTags:     groupByTags,
		GroupByMetaKeys: groupByMetaKeys,
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

func (m *queryMount) listNestedEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	fs.Infof(nil, "Listing nested entries")
	if pathSegment == "" {
		entries := make([]fs.DirEntry, 0, len(m.GroupByMetaKeys)+1)
		entries = append(entries, &entry.VDirEntry{Name: "Tags", FullPath: path.Join(fullpath, "Tags")})
		for _, metaKey := range m.GroupByMetaKeys {
			entries = append(entries, &entry.VDirEntry{Name: metaKey, FullPath: path.Join(fullpath, metaKey)})
		}
		return entries, nil
	}

	splitted := strings.SplitN(pathSegment, "/", 2)
	head := splitted[0]

	rootQuery := payload.NewStructuredQuery(m.Expression).WithSize(0).WithFacets(true) // We're grouping, we don't need any results.
	results, err := m.client.Query(rootQuery)
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
	rootQuery := payload.NewStructuredQuery(m.Expression)
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
	fs.Infof(nil, "listing query entries for '%s'", pathSegment)

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
