package filesystem

import (
	"time"

	"github.com/rclone/rclone/cmd/mountlib"
)

func getMountLibOptions() mountlib.Options {
	return mountlib.Options{
		MaxReadAhead:  128 * 1024,
		AttrTimeout:   1 * time.Second, // how long the kernel caches attribute for
		NoAppleDouble: true,            // use noappledouble by default
		NoAppleXattr:  false,           // do not use noapplexattr by default
		AsyncRead:     true,            // do async reads by default
	}
}
