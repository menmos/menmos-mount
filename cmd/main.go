package main

import (
	"log"
	"os"

	"github.com/menmos/menmos-mount/filesystem"
	"github.com/urfave/cli/v2"
)

func getMountConfig(c *cli.Context) (cfg filesystem.Config, err error) {
	if path := c.String("config"); path != "" {
		cfg, err = LoadConfig(path)
	} else {
		cfg, err = LoadOrCreateDefaultConfig()
	}
	return
}

func initMount(c *cli.Context) error {
	// Load the filesystem configuration.
	cfg, err := getMountConfig(c)
	if err != nil {
		return err
	}

	verbose := c.Bool("verbose")

	mount, err := filesystem.Mount(cfg, verbose)
	if err != nil {
		return err
	}

	return mount.Wait()
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
