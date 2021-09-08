// +build !linux,!darwin,!freebsd

package filesystem

import (
	"github.com/rclone/rclone/vfs/vfscommon"
)

func getVFSOptions() vfscommon.Options {
	return getDefaultVFSOptions()
}
