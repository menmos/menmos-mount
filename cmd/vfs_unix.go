// +build linux darwin freebsd

package main

import (
	"github.com/rclone/rclone/vfs/vfscommon"
	"golang.org/x/sys/unix"
)

func getVFSOptions() vfscommon.Options {
	defaultOpt := getDefaultVFSOptions()

	defaultOpt.Umask = unix.Umask(0) // read the umask
	unix.Umask(defaultOpt.Umask)     // set it back to what it was
	defaultOpt.UID = uint32(unix.Geteuid())
	defaultOpt.GID = uint32(unix.Getegid())

	return defaultOpt
}
