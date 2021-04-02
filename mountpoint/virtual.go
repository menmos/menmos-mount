package mountpoint

import (
	"context"
	"path"
	"strings"

	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type virtualMount struct {
	mounts map[string]MountPoint
}

func NewVirtualMount(mounts map[string]MountPoint) MountPoint {
	return &virtualMount{mounts}
}

func (m *virtualMount) ListEntries(ctx context.Context, pathSegment string, fullpath string) (fs.DirEntries, error) {
	fs.Infof(nil, "listing vmount entries for '%s'", pathSegment)
	splittedPath := strings.SplitN(pathSegment, "/", 2)
	head := splittedPath[0]

	if head == "" || head == "." {
		entries := make([]fs.DirEntry, 0, len(m.mounts))
		for mountName := range m.mounts {
			entries = append(entries, &entry.VDirEntry{Name: mountName, FullPath: path.Join(fullpath, mountName)})
		}
		return entries, nil
	}

	var tail string
	if len(splittedPath) == 2 {
		tail = splittedPath[1]
	} else {
		tail = ""
	}

	if mount, ok := m.mounts[head]; ok {
		return mount.ListEntries(ctx, tail, fullpath)
	}

	return nil, fs.ErrorDirNotFound
}

func (m *virtualMount) ResolveBlobDirectory(path string) (*entry.DirectoryBlobEntry, bool) {
	splittedPath := strings.SplitN(path, "/", 2)
	head := splittedPath[0]

	if head == "" {
		// Virtual mount directories are not blobs.
		return nil, false
	}

	var tail string
	if len(splittedPath) == 2 {
		tail = splittedPath[1]
	} else {
		tail = ""
	}

	if mount, ok := m.mounts[head]; ok {
		return mount.ResolveBlobDirectory(tail)
	}

	return nil, false
}

func (m *virtualMount) ResolveBlobFile(path string) (*entry.FileBlobEntry, bool) {
	fs.Infof(nil, "vmount resolving blob file: %s", path)
	splittedPath := strings.SplitN(path, "/", 2)
	head := splittedPath[0]

	if head == "" {
		// Virtual mount directories are not blobs.
		return nil, false
	}

	var tail string
	if len(splittedPath) == 2 {
		tail = splittedPath[1]
	} else {
		tail = ""
	}

	if mount, ok := m.mounts[head]; ok {
		return mount.ResolveBlobFile(tail)
	}

	return nil, false
}
