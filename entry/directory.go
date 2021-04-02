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
	results, err := b.client.Query(payload.NewStructuredQuery(payload.NewExpression().AndParent(b.BlobID)).WithSize(0)) // With a size of 0 we load no document - query is faster.
	if err != nil {
		// TODO: Log once we have logging.
		return -1
	}

	return int64(results.Total)
}

func (b *DirectoryBlobEntry) ID() string {
	return b.BlobID
}
