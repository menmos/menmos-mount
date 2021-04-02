package filesystem

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"time"

	"github.com/menmos/menmos-go"
	"github.com/menmos/menmos-go/payload"
	"github.com/menmos/menmos-mount/entry"
	"github.com/menmos/menmos-mount/mountpoint"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/hash"
)

// Filesystem provides access to a menmos cluster.
type Filesystem struct {
	name  string
	mount mountpoint.MountPoint

	Client *menmos.Client
}

func NewFs(ctx context.Context, config Config) (fs.Fs, error) {
	client, err := menmos.NewFromProfile(config.Profile)
	if err != nil {
		return nil, err
	}

	f := &Filesystem{
		name:   "menmos",
		Client: client,
	}

	mount, err := mountpoint.Load(config.Mount, client, f)
	if err != nil {
		return nil, err
	}

	f.mount = mount

	return f, nil
}

// Name returns the name of the remote.
func (f *Filesystem) Name() string {
	return f.name
}

// Root returns the mounted filesystem root.
func (f *Filesystem) Root() string {
	return "/"
}

// String returns a description of the FS
func (f *Filesystem) String() string {
	return "Menmos cluster " // TODO: Improve
}

// Precision returns the timestamp precision of the filesystem.
// For Menmos, this is seconds.
func (f *Filesystem) Precision() time.Duration {
	return 1000000000
}

// Hashes returns the supported hash types of this filesystem.
func (f *Filesystem) Hashes() hash.Set {
	return hash.NewHashSet()
}

// Features returns the supported features of this filesystem.
func (f *Filesystem) Features() *fs.Features {
	// TODO: Actually support some features.
	return &fs.Features{
		CaseInsensitive:         false,
		DuplicateFiles:          true,
		ReadMimeType:            true,
		WriteMimeType:           true,
		CanHaveEmptyDirectories: true,
		BucketBased:             false,
		BucketBasedRootOK:       false,
		SetTier:                 false,
		GetTier:                 false,
		ServerSideAcrossConfigs: false,
		IsLocal:                 false,
		SlowModTime:             true,
		SlowHash:                true,

		Move: f.Move,
	}
}

func (f *Filesystem) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	entries, err = f.mount.ListEntries(ctx, dir, dir)
	return
}

func (f *Filesystem) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	// TODO: In normal filesystem use, this isn't called. Not sure this is actually required for our use case.
	// Implement this if it causes problems down the line.
	return nil, fs.ErrorNotImplemented
}

func (f *Filesystem) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	fs.Infof(nil, "received PUT request for %s", src.Remote())
	objectSize := src.Size()
	if objectSize == -1 {
		fs.Infof(nil, "PUT failed")
		return nil, errors.New("object size needs to be known to upload")
	}

	// To put the object, we first need the blob ID of its parent directory.
	// TODO: Put is called for updates AND creations - distinguish the two before uploading.
	if parentDirectory, ok := f.mount.ResolveBlobDirectory(filepath.Dir(src.Remote())); ok {
		fs.Infof(nil, "found parent blob: %s", parentDirectory.BlobID)

		if currentFile, ok := f.mount.ResolveBlobFile(src.Remote()); ok {
			// Update
			currentFile.Meta.Size = uint64(src.Size())
			if err := f.Client.UpdateBlob(currentFile.BlobID, io.NopCloser(in), currentFile.Meta); err != nil {
				return nil, err
			}
			return currentFile, nil
		}
		// Create
		meta := payload.NewBlobMeta(filepath.Base(src.Remote()), "File", uint64(objectSize))
		meta.Parents = append(meta.Parents, parentDirectory.BlobID)
		blobID, err := f.Client.CreateBlob(io.NopCloser(in), meta)
		if err != nil {
			fs.Infof(nil, "PUT failed", err.Error())
			return nil, err
		}
		fs.Infof(nil, "PUT success", blobID)
		return entry.NewFile(blobID, meta, src.Remote(), f.Client, f), nil
	}

	fs.Infof(nil, "permission denied")
	return nil, fs.ErrorPermissionDenied
}

func (f *Filesystem) Mkdir(ctx context.Context, dir string) error {
	fs.Infof(nil, "received MKDIR for: %s")

	if _, fileOk := f.mount.ResolveBlobFile(dir); fileOk {
		return fs.ErrorIsFile
	}

	if _, dirOk := f.mount.ResolveBlobDirectory(dir); dirOk {
		return fs.ErrorDirExists
	}

	if parentDirectory, ok := f.mount.ResolveBlobDirectory(filepath.Dir(dir)); ok {
		fs.Infof(nil, "found new parent blob: %s", parentDirectory.BlobID)
		meta := payload.NewBlobMeta(filepath.Base(dir), "Directory", 0)
		meta.Parents = append(meta.Parents, parentDirectory.BlobID)
		blobID, err := f.Client.CreateBlob(nil, meta)
		if err != nil {
			return err
		}

		fs.Infof(nil, "PUT success: ", blobID)
		return nil
	}

	return fs.ErrorPermissionDenied
}

func (f *Filesystem) Rmdir(ctx context.Context, dir string) error {
	parentEntry, ok := f.mount.ResolveBlobDirectory(dir)
	if !ok {
		return fs.ErrorDirNotFound
	}

	// Make sure the directory is empty.
	response, err := f.Client.Query(payload.NewStructuredQuery(payload.NewExpression().AndParent(parentEntry.BlobID)).WithSize(0))
	if err != nil {
		return err
	}

	if response.Total > 0 {
		return fs.ErrorDirectoryNotEmpty
	}

	return f.Client.Delete(parentEntry.BlobID)
}

func (f *Filesystem) Move(ctx context.Context, src fs.Object, remote string) (fs.Object, error) {
	srcParentDir, ok := f.mount.ResolveBlobDirectory(filepath.Dir(src.Remote()))
	if !ok {
		return nil, fs.ErrorCantMove
	}

	if srcFile, ok := f.mount.ResolveBlobFile(src.Remote()); ok {
		// Delete destination path if it exists.
		if dstFile, ok := f.mount.ResolveBlobFile(remote); ok {
			if err := dstFile.Remove(ctx); err != nil {
				// TODO: If delete fails, what should we do here?
				return nil, fs.ErrorCantMove
			}
		}

		if dstParentDir, ok := f.mount.ResolveBlobDirectory(filepath.Dir(remote)); ok {
			newParents := make([]string, 0, len(srcFile.Meta.Parents))
			newParents = append(newParents, dstParentDir.ID())
			for _, parentID := range srcFile.Meta.Parents {
				if parentID != srcParentDir.ID() {
					newParents = append(newParents, parentID)
				}
			}
			srcFile.Meta.Parents = newParents
			srcFile.Meta.Name = filepath.Base(remote)

			if err := f.Client.UpdateMeta(srcFile.BlobID, srcFile.Meta); err != nil {
				// TODO: Log
				return nil, err
			}
			return srcFile, nil
		}
	}
	return nil, fs.ErrorCantMove
}
