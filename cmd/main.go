package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/menmos/menmos-mount/filesystem"
	_ "github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/vfs"
)

func loadConfig(path string) (filesystem.Config, error) {
	rawCfg, err := os.ReadFile(path)
	if err != nil {
		return filesystem.Config{}, err
	}

	var cfg filesystem.Config
	err = json.Unmarshal(rawCfg, &cfg)

	return cfg, err
}

func main() {
	cfg, err := loadConfig("./cfg.json")
	if err != nil {
		panic(err)
	}

	configfile.LoadConfig(context.Background())

	fs, err := filesystem.NewFs(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	opt := getVFSOptions()
	vfs := vfs.New(fs, &opt)

	// TODO: Read mount point from CLI args instead.
	if err := mountlib.Mount(vfs, cfg.Mountpoint, doMount, nil); err != nil {
		panic(err)
	}
}
