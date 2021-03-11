package main

import "github.com/rclone/rclone/vfs/vfscommon"

func getDefaultVFSOptions() vfscommon.Options {
	opt := vfscommon.DefaultOpt
	opt.CacheMode = vfscommon.CacheModeWrites

	return opt
}
