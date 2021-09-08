// +build linux darwin freebsd

package filesystem

import (
	"github.com/rclone/rclone/vfs/vfscommon"
	"golang.org/x/sys/unix"
)

func getVFSOptions() vfscommon.Options {
	defaultOpt := getDefaultVFSOptions()

	// We overwrite the umask with 0 to get the _previous_ umask (its a hack to read the mask), then we set it back to the previous value
	// so nothign changes in the end.
	defaultOpt.Umask = unix.Umask(0) // read the umask
	unix.Umask(defaultOpt.Umask)     // set it back to what it was
	defaultOpt.UID = uint32(unix.Geteuid())
	defaultOpt.GID = uint32(unix.Getegid())

	return defaultOpt
}
