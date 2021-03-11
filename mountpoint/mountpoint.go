package mountpoint

import (
	"context"

	"github.com/menmos/menmos-mount/entry"
	"github.com/rclone/rclone/fs"
)

type MountPoint interface {
	ListEntries(ctx context.Context, path string, fullpath string) (fs.DirEntries, error)
	ResolveBlobDirectory(path string) (*entry.DirectoryBlobEntry, bool)
	ResolveBlobFile(path string) (*entry.FileBlobEntry, bool)
}
