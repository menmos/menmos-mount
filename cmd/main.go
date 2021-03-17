package main

import (
	"context"
	"log"
	"os"

	"github.com/menmos/menmos-mount/filesystem"
	_ "github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configfile"
	"github.com/rclone/rclone/vfs"
	"github.com/urfave/cli/v2"
)

func getMountConfig(c *cli.Context) (cfg filesystem.Config, err error) {
	if path := c.String("config"); path != "" {
		cfg, err = LoadConfig(path)
	} else {
		cfg, err = LoadDefaultConfig()
	}
	return
}

func initRcloneEnv(c *cli.Context) {
	configfile.LoadConfig(context.Background())

	if c.Bool("verbose") {
		rcloneConfig := fs.GetConfig(context.Background())
		rcloneConfig.LogLevel = fs.LogLevelDebug
	}
}

func initMount(c *cli.Context) error {
	// Do some work to setup rclone prereqs.
	initRcloneEnv(c)

	// Load the filesystem configuration.
	cfg, err := getMountConfig(c)
	if err != nil {
		return err
	}

	// Create a menmos filesystem from our config.
	fs, err := filesystem.NewFs(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	// Configure a VFS for our filesystem.
	opt := getVFSOptions()
	vfs := vfs.New(fs, &opt)

	// Mount.
	return mountlib.Mount(vfs, cfg.Mountpoint, doMount, nil)
}

func main() {
	app := &cli.App{
		Name:  "menmos-mount",
		Usage: "Filesystem Interface to Menmos",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "",
				Usage: "the config file to use (defaults to $CONFIG_DIR/menmos/mount.json)",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Value: false,
				Usage: "whether to use verbose output",
			},
		},
	}

	app.Action = initMount

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
