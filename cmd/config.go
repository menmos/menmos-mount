package main

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/menmos/menmos-mount/filesystem"
	"github.com/pkg/errors"
)

const menmosConfigDirName = "menmos"
const mountConfigFileName = "mount.json"

func getDefaultConfigPath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get the user configuration directory")
	}

	menmosConfigPath := path.Join(configPath, menmosConfigDirName, mountConfigFileName)
	return menmosConfigPath, nil
}

func LoadConfig(path string) (filesystem.Config, error) {
	var cfg filesystem.Config

	rawCfg, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(rawCfg, &cfg)
	return cfg, err
}

func LoadOrCreateDefaultConfig() (filesystem.Config, error) {
	path, err := getDefaultConfigPath()
	if err != nil {
		return filesystem.Config{}, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0644); err != nil {
		return filesystem.Config{}, errors.Wrap(err, "failed to create menmos config directory")
	}

	return LoadConfig(path)
}
