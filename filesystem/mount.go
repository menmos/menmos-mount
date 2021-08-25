package filesystem

import (
	"context"

	_ "github.com/rclone/rclone/backend/local"

	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
)

func initRcloneEnvironment(verbose bool) {
	if verbose {
		rcloneConfig := fs.GetConfig(context.Background())
		rcloneConfig.LogLevel = fs.LogLevelDebug
	}
}

func Mount(config Config, verbose bool) (*mountlib.MountPoint, error) {
	initRcloneEnvironment(verbose)

	fs, err := NewFs(context.Background(), config)
	if err != nil {
		return nil, err
	}

	vfsOptions := getVFSOptions()
	mountOptions := getMountLibOptions()

	mount := &mountlib.MountPoint{
		MountFn:    doMount,
		MountPoint: config.Mountpoint,
		Fs:         fs,
		MountOpt:   mountOptions,
		VFSOpt:     vfsOptions,
	}

	if _, err := mount.Mount(); err != nil {
		return nil, err
	}

	return mount, nil
}
