package entry

import (
	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/rclone/rclone/fs"
)

type DirectoryBlobEntry struct {
	BlobEntry
}

func NewDirectory(blobID string, blobMeta payload.BlobMeta, path string, client *menmos.Client, fs fs.Info) *DirectoryBlobEntry {
	return &DirectoryBlobEntry{BlobEntry: BlobEntry{
		BlobID: blobID,
		Meta:   blobMeta,
		path:   path,
		client: client,
		fs:     fs,
	}}
}

func (b *DirectoryBlobEntry) Items() int64 {
	// TODO: Run a query to find out the nb. of children of given blob.
	return -1
}

func (b *DirectoryBlobEntry) ID() string {
	return b.BlobID
}
