// +build !linux,!darwin,!freebsd

package main

import (
	"github.com/rclone/rclone/vfs/vfscommon"
)

func getVFSOptions() vfscommon.Options {
	return getDefaultVFSOptions()
}
