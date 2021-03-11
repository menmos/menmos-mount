package entry

import (
	"context"
	"time"
)

type VDirEntry struct {
	Name     string
	FullPath string
}

func (e *VDirEntry) String() string {
	return e.Name
}

func (e *VDirEntry) Remote() string {
	return e.FullPath
}

func (e *VDirEntry) ModTime(context.Context) time.Time {
	return time.Unix(0, 0)
}

func (e *VDirEntry) Size() int64 {
	return 0
}

func (e *VDirEntry) Items() int64 {
	return -1
}

func (e *VDirEntry) ID() string {
	return ""
}
