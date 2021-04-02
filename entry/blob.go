package entry

import (
	"context"
	"time"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/rclone/rclone/fs"
)

type BlobEntry struct {
	BlobID string
	Meta   payload.BlobMeta

	path   string
	client *menmos.Client
	fs     fs.Info
}

func (e *BlobEntry) String() string {
	return e.BlobID
}

func (e *BlobEntry) Remote() string {
	return e.path
}

func (e *BlobEntry) ModTime(context.Context) time.Time {
	return time.Unix(0, 0)
}

func (e *BlobEntry) Size() int64 {
	return int64(e.Meta.Size) // Technically not super safe to cast u64 to i64, but this will only fail on files above 9200 Pb... We should be OK.
}
