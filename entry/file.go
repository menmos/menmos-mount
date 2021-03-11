package entry

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/hash"
)

type FileBlobEntry struct {
	BlobEntry
}

func NewFile(blobID string, blobMeta payload.BlobMeta, path string, client *menmos.Client, fs fs.Info) *FileBlobEntry {
	return &FileBlobEntry{BlobEntry: BlobEntry{
		BlobID: blobID,
		Meta:   blobMeta,
		path:   path,
		client: client,
		fs:     fs,
	}}
}

func (b *FileBlobEntry) Fs() fs.Info {
	return b.fs
}

func (b *FileBlobEntry) Hash(ctx context.Context, ty hash.Type) (string, error) {
	// TODO: Add hashing here if menmos supports it one day.
	return "", nil
}

func (b *FileBlobEntry) Storable() bool {
	return false
}

func (b *FileBlobEntry) SetModTime(ctx context.Context, t time.Time) error {
	// TODO: implement once menmos supports mod times.
	return nil
}

func (b *FileBlobEntry) Open(ctx context.Context, options ...fs.OpenOption) (io.ReadCloser, error) {
	fs.FixRangeOption(options, int64(b.BlobEntry.Meta.Size))

	var rangeStart int64 = 0
	var rangeEnd int64 = b.Size() - 1

	for _, option := range options {
		if r, ok := option.(*fs.RangeOption); ok {
			rangeStart = r.Start
			rangeEnd = r.End
			break
		} else if seek, ok := option.(*fs.SeekOption); ok {
			rangeStart = seek.Offset
			rangeEnd = b.Size() - 1 - seek.Offset
			break
		} else if option.Mandatory() {
			return nil, fmt.Errorf("unhandled option: %s", option.String())
		}
	}

	// FixRangeOption() sets the range end to -1 to indicate "unbounded".
	if rangeEnd <= 0 {
		rangeEnd = b.Size() - 1
	}

	return b.client.Get(b.BlobID, &menmos.Range{Start: rangeStart, End: rangeEnd})
}

func (b *FileBlobEntry) Update(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) error {
	for _, option := range options {
		if _, ok := option.(*fs.RangeOption); ok {
			fmt.Println("GOT RANGE OPTION")
			break
		} else if _, ok := option.(*fs.SeekOption); ok {
			fmt.Println("GOT SEEK OPTION")
			break
		} else if option.Mandatory() {
			return fmt.Errorf("unhandled option: %s", option.String())
		}
	}

	return fs.ErrorNotImplemented
}

func (b *FileBlobEntry) Remove(ctx context.Context) error {
	if b.BlobID == "" {
		fmt.Println("DELETE - NO BLOB ID DEFINED: ", *b)
		return nil
	}

	return b.client.Delete(b.BlobID)
}
